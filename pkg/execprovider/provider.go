// Package execprovider implements a generic C2 provider that wraps external shell commands.
// Each instance is configured via a YAML file in ~/.c4/providers/<name>.yaml.
package execprovider

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/CR1MS0N-Operator/c4/pkg/provider"
	"gopkg.in/yaml.v3"
)

// HealthConfig describes how to check liveness of the C2 process.
type HealthConfig struct {
	HTTP string `yaml:"http"` // HTTP(S) URL to GET
	Cmd  string `yaml:"cmd"`  // shell command to run (exit 0 = healthy)
}

// ExecConfig is the top-level YAML structure for an exec provider.
type ExecConfig struct {
	Type     string       `yaml:"type"`
	Name     string       `yaml:"name"`
	StartCmd string       `yaml:"start"`
	StopCmd  string       `yaml:"stop"`
	Health   HealthConfig `yaml:"health"`
	LogPath  string       `yaml:"log_path"`
	Timeout  int          `yaml:"timeout"` // seconds
}

// Provider implements provider.Provider by wrapping external commands.
type Provider struct {
	name   string
	cfg    ExecConfig
	health provider.HealthCheckResult
}

// New creates a new exec provider from a parsed config.
func New(cfg ExecConfig) *Provider {
	return &Provider{name: cfg.Name, cfg: cfg}
}

// Load reads a YAML file and returns an ExecConfig.
func Load(path string) (*ExecConfig, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg ExecConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if cfg.Type != "exec" {
		return nil, fmt.Errorf("%s: type must be \"exec\", got %q", path, cfg.Type)
	}
	if cfg.Name == "" {
		return nil, fmt.Errorf("%s: name is required", path)
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30
	}
	return &cfg, nil
}

// LoadDir scans a directory for *.yaml files and loads each as an exec config.
func LoadDir(dir string) ([]ExecConfig, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	var configs []ExecConfig
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		cfg, err := Load(filepath.Join(dir, e.Name()))
		if err != nil {
			return configs, fmt.Errorf("loading %s: %w", e.Name(), err)
		}
		configs = append(configs, *cfg)
	}
	return configs, nil
}

// Type returns the provider type (Exec).
func (p *Provider) Type() provider.Type {
	return provider.TypeExec
}

// Name returns the provider instance name.
func (p *Provider) Name() string {
	return p.name
}

// Connect validates that the configured start command binary exists.
func (p *Provider) Connect(_ context.Context) error {
	if p.cfg.StartCmd == "" {
		return fmt.Errorf("%s: start command is empty", p.name)
	}
	binary := strings.Fields(p.cfg.StartCmd)[0]
	if _, err := exec.LookPath(binary); err != nil {
		return fmt.Errorf("%s: start command %q not found in PATH: %w", p.name, binary, err)
	}
	return nil
}

// Disconnect is a no-op for exec providers.
func (p *Provider) Disconnect(_ context.Context) error {
	return nil
}

// Deploy runs the start command with a timeout.
func (p *Provider) Deploy(ctx context.Context) error {
	timeout := time.Duration(p.cfg.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", p.cfg.StartCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s start failed: %w", p.name, err)
	}
	return nil
}

// Destroy runs the stop command with a timeout.
func (p *Provider) Destroy(ctx context.Context) error {
	if p.cfg.StopCmd == "" {
		return fmt.Errorf("%s: stop command is empty", p.name)
	}
	timeout := time.Duration(p.cfg.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", p.cfg.StopCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s stop failed: %w", p.name, err)
	}
	return nil
}

// Health checks whether the C2 process is alive.
func (p *Provider) Health(ctx context.Context) (*provider.HealthCheckResult, error) {
	// Prefer HTTP health check
	if p.cfg.Health.HTTP != "" {
		return p.httpHealth(ctx)
	}
	// Fallback to command-based health check
	if p.cfg.Health.Cmd != "" {
		return p.cmdHealth(ctx)
	}
	return &provider.HealthCheckResult{
		Healthy:   false,
		Message:   "no health check configured (set health.http or health.cmd)",
		Timestamp: time.Now(),
	}, nil
}

func (p *Provider) httpHealth(ctx context.Context) (*provider.HealthCheckResult, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.Health.HTTP, nil)
	if err != nil {
		return &provider.HealthCheckResult{
			Healthy: false, Message: fmt.Sprintf("bad health URL: %s", err.Error()),
			Timestamp: time.Now(),
		}, nil
	}
	resp, err := client.Do(req)
	if err != nil {
		return &provider.HealthCheckResult{
			Healthy: false, Message: fmt.Sprintf("health check failed: %s", err.Error()),
			Timestamp: time.Now(),
		}, nil
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return &provider.HealthCheckResult{
			Healthy: true, Message: fmt.Sprintf("HTTP %d from %s", resp.StatusCode, p.cfg.Health.HTTP),
			Timestamp: time.Now(),
		}, nil
	}
	return &provider.HealthCheckResult{
		Healthy: false, Message: fmt.Sprintf("unexpected HTTP %d from %s", resp.StatusCode, p.cfg.Health.HTTP),
		Timestamp: time.Now(),
	}, nil
}

func (p *Provider) cmdHealth(_ context.Context) (*provider.HealthCheckResult, error) {
	cmd := exec.Command("sh", "-c", p.cfg.Health.Cmd)
	if err := cmd.Run(); err != nil {
		return &provider.HealthCheckResult{
			Healthy: false, Message: fmt.Sprintf("health command exited non-zero: %s", err.Error()),
			Timestamp: time.Now(),
		}, nil
	}
	return &provider.HealthCheckResult{
		Healthy: true, Message: "health command succeeded",
		Timestamp: time.Now(),
	}, nil
}

// Stubs for unsupported operations.

func (p *Provider) Listeners(_ context.Context) ([]provider.Listener, error) {
	return []provider.Listener{}, fmt.Errorf("listeners not supported for exec providers")
}
func (p *Provider) StartListener(_ context.Context, _ provider.Listener) (*provider.Listener, error) {
	return nil, fmt.Errorf("start-listener not supported for exec providers")
}
func (p *Provider) StopListener(_ context.Context, _ string) error {
	return fmt.Errorf("stop-listener not supported for exec providers")
}
func (p *Provider) Callbacks(_ context.Context) ([]provider.Callback, error) {
	return []provider.Callback{}, fmt.Errorf("callbacks not supported for exec providers")
}
func (p *Provider) Payloads(_ context.Context) ([]provider.Payload, error) {
	return []provider.Payload{}, fmt.Errorf("payloads not supported for exec providers")
}
func (p *Provider) GeneratePayload(_ context.Context, _ map[string]any) (*provider.Payload, error) {
	return nil, fmt.Errorf("generate-payload not supported for exec providers")
}

// Ensure Provider implements provider.Provider.
var _ provider.Provider = (*Provider)(nil)
