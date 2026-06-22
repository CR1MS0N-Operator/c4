// Package mythic implements the Mythic C2 provider interface.
package mythic

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/CR1MS0N-Operator/c4/pkg/config"
	"github.com/CR1MS0N-Operator/c4/pkg/graphql"
	"github.com/CR1MS0N-Operator/c4/pkg/provider"
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
// callbackResponse is used to unmarshal the callback query.
type callbackResponse struct {
	Data struct {
		Callbacks []mythicCallback `json:"callback"`
	} `json:"data"`
}

type mythicCallback struct {
	ID             int    `json:"id"`
	DisplayID      int    `json:"display_id"`
	IP             string `json:"ip"`
	Host           string `json:"host"`
	User           string `json:"user"`
	ProcessName    string `json:"process_name"`
	PID            int    `json:"pid"`
	OS             string `json:"os"`
	Architecture   string `json:"architecture"`
	LastCheckin    string `json:"last_checkin"`
	SleepInfo      string `json:"sleep_info"`
	Description    string `json:"description"`
	Domain         string `json:"domain"`
}

// payloadResponse is used to unmarshal the payload query.
type payloadResponse struct {
	Data struct {
		Payloads []mythicPayload `json:"payload"`
	} `json:"data"`
}

type mythicPayload struct {
	ID          int    `json:"id"`
	UUID        string `json:"uuid"`
	Description string `json:"description"`
	AutoGen     bool   `json:"auto_generated"`
	BuildPhase  string `json:"build_phase"`
	PayloadTypeSemver string `json:"payload_type_semver"`
	OS          string `json:"os"`
	CreationTime string `json:"creation_time"`
}

// Callbacks returns active callbacks from Mythic.
func (p *MythicProvider) Callbacks(ctx context.Context) ([]provider.Callback, error) {
	if p.client == nil {
		return []provider.Callback{}, fmt.Errorf("not connected — call Connect first")
	}

	var resp callbackResponse
	if err := p.client.Do(ctx, `{ callback(where: {active: {_eq: true}}, order_by: {last_checkin: desc}) { id display_id ip host user process_name pid os architecture last_checkin sleep_info description domain } }`, &resp); err != nil {
		return []provider.Callback{}, fmt.Errorf("list callbacks: %w", err)
	}

	callbacks := make([]provider.Callback, 0, len(resp.Data.Callbacks))
	for _, c := range resp.Data.Callbacks {
		callbacks = append(callbacks, provider.Callback{
			ID:       strconv.Itoa(c.ID),
			Listener: c.Description,
			Agent:    fmt.Sprintf("%s/%s", c.OS, c.Architecture),
			Host:     c.Host,
			User:     c.User,
			PID:      c.PID,
			Status:   provider.StatusRunning,
		})
	}

	return callbacks, nil
}

// Payloads returns payloads known to Mythic.
func (p *MythicProvider) Payloads(ctx context.Context) ([]provider.Payload, error) {
	if p.client == nil {
		return []provider.Payload{}, fmt.Errorf("not connected — call Connect first")
	}

	var resp payloadResponse
	if err := p.client.Do(ctx, `{ payload { id uuid description auto_generated build_phase payload_type_semver os creation_time } }`, &resp); err != nil {
		return []provider.Payload{}, fmt.Errorf("list payloads: %w", err)
	}

	payloads := make([]provider.Payload, 0, len(resp.Data.Payloads))
	for _, p2 := range resp.Data.Payloads {
		payloads = append(payloads, provider.Payload{
			ID:       strconv.Itoa(p2.ID),
			Name:     p2.Description,
			Type:     p2.PayloadTypeSemver,
			Format:   p2.BuildPhase,
			FilePath: p2.UUID,
		})
	}

	return payloads, nil
}
