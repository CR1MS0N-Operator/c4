// Package mythic implements the Mythic C2 provider interface.
package mythic

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ForeverLX/c4/pkg/config"
	"github.com/ForeverLX/c4/pkg/graphql"
	"github.com/ForeverLX/c4/pkg/provider"
)

// MythicProvider implements the provider.Provider interface for Mythic.
type MythicProvider struct {
	name   string
	cfg    config.C2Config
	client *graphql.Client
}

// NewMythicProvider creates a new Mythic provider instance.
func NewMythicProvider(name string, cfg config.C2Config) *MythicProvider {
	return &MythicProvider{name: name, cfg: cfg}
}

// Type returns the provider type.
func (p *MythicProvider) Type() provider.Type {
	return provider.TypeMythic
}

// Name returns the provider instance name.
func (p *MythicProvider) Name() string {
	return p.name
}

// endpoint returns the GraphQL URL for the given config.
func endpoint(cfg config.C2Config) string {
	scheme := "http"
	if cfg.SSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d/graphql/", scheme, cfg.Host, cfg.APIPort)
}

// Connect establishes a connection to the Mythic server.
// For Mythic v3 (Hasura-backed), it uses the x-hasura-admin-secret for auth.
func (p *MythicProvider) Connect(ctx context.Context) error {
	ep := endpoint(p.cfg)

	// Try primary endpoint first (nginx on api_port)
	client, err := p.tryConnect(ctx, ep, p.cfg.HasuraSecret)
	if err == nil {
		p.client = client
		return nil
	}

	// If primary fails and we have a server_port, try direct backend as fallback
	if p.cfg.ServerPort > 0 && p.cfg.ServerPort != p.cfg.APIPort {
		fallbackEp := fmt.Sprintf("http://%s:%d/graphql/", p.cfg.Host, p.cfg.ServerPort)
		fallbackClient, fbErr := p.tryConnect(ctx, fallbackEp, p.cfg.HasuraSecret)
		if fbErr == nil {
			p.client = fallbackClient
			return nil
		}
		return fmt.Errorf("primary (%s): %w; fallback (%s): %w", ep, err, fallbackEp, fbErr)
	}

	return fmt.Errorf("connect to %s: %w", ep, err)
}

// tryConnect attempts to create a client and verify connectivity.
func (p *MythicProvider) tryConnect(ctx context.Context, endpoint, secret string) (*graphql.Client, error) {
	timeout := 10 * time.Second
	client := graphql.NewClient(endpoint, secret, true, timeout)

	// Verify connectivity with a simple query
	var resp struct {
		Data struct {
			Operator []struct {
				ID                 int    `json:"id"`
				Username           string `json:"username"`
				CurrentOperationID int    `json:"current_operation_id"`
			} `json:"operator"`
		} `json:"data"`
	}
	if err := client.Do(ctx, `{ operator { id username current_operation_id } }`, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data.Operator) == 0 {
		return nil, fmt.Errorf("authenticated but no operators found")
	}
	return client, nil
}

// Disconnect closes the connection to the Mythic server.
func (p *MythicProvider) Disconnect(_ context.Context) error {
	p.client = nil
	return nil
}

// Deploy deploys the Mythic instance via Docker Compose.
func (p *MythicProvider) Deploy(_ context.Context) error {
	return fmt.Errorf("not yet implemented")
}

// Destroy tears down the Mythic instance.
func (p *MythicProvider) Destroy(_ context.Context) error {
	return fmt.Errorf("not yet implemented")
}

// meResponse is used to unmarshal the operator health-check query.
type meResponse struct {
	Data struct {
		Operator []struct {
			ID                 int    `json:"id"`
			Username           string `json:"username"`
			CurrentOperationID int    `json:"current_operation_id"`
		} `json:"operator"`
	} `json:"data"`
}

// Health checks whether the Mythic server is reachable and healthy.
// Returns a HealthCheckResult with operator info on success.
func (p *MythicProvider) Health(ctx context.Context) (*provider.HealthCheckResult, error) {
	if p.client == nil {
		return &provider.HealthCheckResult{
			Healthy:   false,
			Message:   "not connected — call Connect first",
			Timestamp: time.Now(),
		}, nil
	}

	var resp meResponse
	if err := p.client.Do(ctx, `{ operator { id username current_operation_id } }`, &resp); err != nil {
		return &provider.HealthCheckResult{
			Healthy:   false,
			Message:   fmt.Sprintf("health check failed: %s", err.Error()),
			Timestamp: time.Now(),
		}, nil
	}

	if len(resp.Data.Operator) == 0 {
		return &provider.HealthCheckResult{
			Healthy:   false,
			Message:   "authenticated but no operators found",
			Timestamp: time.Now(),
		}, nil
	}

	op := resp.Data.Operator[0]
	return &provider.HealthCheckResult{
		Healthy:   true,
		Message:   fmt.Sprintf("connected as %s (operation %d)", op.Username, op.CurrentOperationID),
		Timestamp: time.Now(),
	}, nil
}

type c2profileResponse struct {
	Data struct {
		C2Profiles []c2profile `json:"c2profile"`
	} `json:"data"`
}

type c2profile struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Running         bool   `json:"running"`
	ContainerRunning bool  `json:"container_running"`
	Deleted         bool   `json:"deleted"`
	Semver          string `json:"semver"`
}

// Listeners returns all non-deleted C2 profiles from Mythic.
func (p *MythicProvider) Listeners(ctx context.Context) ([]provider.Listener, error) {
	if p.client == nil {
		return []provider.Listener{}, fmt.Errorf("not connected — call Connect first")
	}

	var resp c2profileResponse
	if err := p.client.Do(ctx, `{ c2profile(where: {deleted: {_eq: false}}) { id name running container_running description semver } }`, &resp); err != nil {
		return []provider.Listener{}, fmt.Errorf("list listeners: %w", err)
	}

	listeners := make([]provider.Listener, 0, len(resp.Data.C2Profiles))
	for _, cp := range resp.Data.C2Profiles {
		status := provider.StatusStopped
		if cp.Running {
			status = provider.StatusRunning
		}
		listeners = append(listeners, provider.Listener{
			ID:       strconv.Itoa(cp.ID),
			Name:     cp.Name,
			Type:     cp.Semver,
			Host:     p.cfg.Host,
			Port:     p.cfg.APIPort,
			Status:   status,
			Provider: provider.TypeMythic,
		})
	}

	return listeners, nil
}

// StartListener starts a C2 profile by name (upsert: if exists, start; if new, create).
func (p *MythicProvider) StartListener(ctx context.Context, l provider.Listener) (*provider.Listener, error) {
	if p.client == nil {
		return nil, fmt.Errorf("not connected — call Connect first")
	}

	// Look up existing profile by name
	var lookup struct {
		Data struct {
			C2Profiles []c2profile `json:"c2profile"`
		} `json:"data"`
	}
	if err := p.client.Do(ctx, fmt.Sprintf(`{ c2profile(where: {name: {_eq: "%s"}}) { id name running } }`, l.Name), &lookup); err != nil {
		return nil, fmt.Errorf("lookup listener %q: %w", l.Name, err)
	}

	var updated c2profile

	if len(lookup.Data.C2Profiles) > 0 {
		existing := lookup.Data.C2Profiles[0]
		// Update existing — set running = true
		var updResp struct {
			Data struct {
				Result c2profile `json:"update_c2profile_by_pk"`
			} `json:"data"`
		}
		if err := p.client.Do(ctx, fmt.Sprintf(`mutation { update_c2profile_by_pk(pk_columns: {id: %d}, _set: {running: true}) { id name running } }`, existing.ID), &updResp); err != nil {
			return nil, fmt.Errorf("start listener %q: %w", l.Name, err)
		}
		updated = updResp.Data.Result
	} else {
		// Create new profile
		desc := l.Name
		if l.Name != "" {
			desc = fmt.Sprintf("Managed by C4 — %s", l.Name)
		}
		var insResp struct {
			Data struct {
				Result c2profile `json:"insert_c2profile_one"`
			} `json:"data"`
		}
		if err := p.client.Do(ctx, fmt.Sprintf(`mutation { insert_c2profile_one(object: {name: "%s", description: "%s"}) { id name running } }`, l.Name, desc), &insResp); err != nil {
			// GraphQL escape common issues — strip quotes in name/desc
			safeName := strings.ReplaceAll(l.Name, `"`, `\"`)
			safeDesc := strings.ReplaceAll(desc, `"`, `\"`)
			if err := p.client.Do(ctx, fmt.Sprintf(`mutation { insert_c2profile_one(object: {name: "%s", description: "%s"}) { id name running } }`, safeName, safeDesc), &insResp); err != nil {
				return nil, fmt.Errorf("create listener %q: %w", l.Name, err)
			}
		}
		updated = insResp.Data.Result
	}

	return &provider.Listener{
		ID:       strconv.Itoa(updated.ID),
		Name:     updated.Name,
		Type:     "Mythic",
		Host:     p.cfg.Host,
		Port:     p.cfg.APIPort,
		Status:   provider.StatusRunning,
		Provider: provider.TypeMythic,
	}, nil
}

// StopListener stops a Mythic C2 profile by ID.
func (p *MythicProvider) StopListener(ctx context.Context, id string) error {
	if p.client == nil {
		return fmt.Errorf("not connected — call Connect first")
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid listener id %q: %w", id, err)
	}

	var resp struct {
		Data struct {
			Result c2profile `json:"update_c2profile_by_pk"`
		} `json:"data"`
	}
	if err := p.client.Do(ctx, fmt.Sprintf(`mutation { update_c2profile_by_pk(pk_columns: {id: %d}, _set: {running: false}) { id name running } }`, intID), &resp); err != nil {
		return fmt.Errorf("stop listener %s: %w", id, err)
	}

	if resp.Data.Result.ID == 0 {
		return fmt.Errorf("listener %s not found", id)
	}

	return nil
}

// Callbacks returns active callbacks from Mythic.
func (p *MythicProvider) Callbacks(_ context.Context) ([]provider.Callback, error) {
	return nil, fmt.Errorf("not yet implemented")
}

// Payloads returns payloads known to Mythic.
func (p *MythicProvider) Payloads(_ context.Context) ([]provider.Payload, error) {
	return nil, fmt.Errorf("not yet implemented")
}

// GeneratePayload generates a new Mythic payload.
func (p *MythicProvider) GeneratePayload(_ context.Context, _ map[string]any) (*provider.Payload, error) {
	return nil, fmt.Errorf("not yet implemented")
}

// Ensure MythicProvider implements provider.Provider.
var _ provider.Provider = (*MythicProvider)(nil)
