// Package detect auto-discovers installed C2 frameworks on the local system.
//
// Result contract: every function in this package that returns []DetectionResult
// returns a non-nil slice. An empty slice means "nothing was found"; nil is
// never a valid return value. Callers depend on this — json.Marshal of a nil
// slice yields `null` instead of `[]`, and `c4 detect --json` must emit an
// empty JSON array rather than null.
//
// Sub-checks are aggregated by RunAll, where a nil result is indistinguishable
// from an empty one, so no sub-check signals failure by returning nil. A check
// that cannot run (missing directory, tool not installed, command failed)
// reports zero results, the same as a scan that ran and found nothing. If a
// check ever needs to surface an unrecoverable failure, give it an `error`
// return value instead of reusing nil as a signal.
package detect

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/CR1MS0N-Operator/c4/pkg/provider"
)

// DetectionResult describes a C2 framework discovered on the system.
type DetectionResult struct {
	Name       string       `json:"name"`
	Type       provider.Type `json:"type"`
	Path       string       `json:"path,omitempty"`
	Running    bool         `json:"running"`
	Port       int          `json:"port,omitempty"`
	Confidence float64      `json:"confidence"` // 0.0–1.0
}

// Runner runs the detection scans and returns results.
type Runner struct {
	homeDir string
}

// New creates a new detection runner.
func New() *Runner {
	home, _ := os.UserHomeDir()
	return &Runner{homeDir: home}
}

// RunAll runs all detection strategies and returns aggregated results.
// The returned slice is always non-nil, even when nothing was detected.
func (r *Runner) RunAll(ctx context.Context) []DetectionResult {
	results := make([]DetectionResult, 0)
	results = append(results, r.checkInstallPaths()...)
	results = append(results, r.checkDocker(ctx)...)
	results = append(results, r.checkProcesses(ctx)...)
	results = append(results, r.checkExecProviders()...)
	return results
}

// checkInstallPaths looks for C2 install directories in ~/Tools/c2/.
// Returns an empty (non-nil) slice when the tools directory does not exist
// or holds no C2 installs — a missing directory is "nothing found", not a failure.
func (r *Runner) checkInstallPaths() []DetectionResult {
	toolsDir := filepath.Join(r.homeDir, "Tools", "c2")
	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		return make([]DetectionResult, 0)
	}

	results := make([]DetectionResult, 0)

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		path := filepath.Join(toolsDir, name)

		switch name {
		case "mythic":
			results = append(results, DetectionResult{
				Name:       "mythic",
				Type:       provider.TypeMythic,
				Path:       path,
				Running:    false, // checked separately
				Confidence: 0.9,
			})
		case "sliver":
			results = append(results, DetectionResult{
				Name:       "sliver",
				Type:       provider.TypeSliver,
				Path:       path,
				Running:    false,
				Confidence: 0.8,
			})
		case "havoc":
			results = append(results, DetectionResult{
				Name:       "havoc",
				Type:       provider.TypeHavoc,
				Path:       path,
				Running:    false,
				Confidence: 0.8,
			})
		default:
			// Generic C2 detected — could have a docker-compose.yml or start script
			if hasFile(path, "docker-compose.yml") || hasFile(path, "start.sh") {
				results = append(results, DetectionResult{
					Name:       name,
					Type:       provider.TypeExec,
					Path:       path,
					Running:    false,
					Confidence: 0.5,
				})
			}
		}
	}
	return results
}

// checkDocker looks for running C2 containers.
// Returns an empty (non-nil) slice when docker is unavailable or the query
// fails — an unreachable daemon means "no containers seen", not a failure.
func (r *Runner) checkDocker(ctx context.Context) []DetectionResult {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "ps", "--format", "{{.Names}}")
	out, err := cmd.Output()
	if err != nil {
		return make([]DetectionResult, 0)
	}

	results := make([]DetectionResult, 0)
	seen := map[string]bool{}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "mythic_"):
			if !seen["mythic"] {
				results = append(results, DetectionResult{
					Name: "mythic", Type: provider.TypeMythic,
					Running: true, Port: 7443, Confidence: 0.95,
				})
				seen["mythic"] = true
			}
		case strings.HasPrefix(line, "sliver_"):
			if !seen["sliver"] {
				results = append(results, DetectionResult{
					Name: "sliver", Type: provider.TypeSliver,
					Running: true, Confidence: 0.7,
				})
				seen["sliver"] = true
			}
		case strings.HasPrefix(line, "havoc_"):
			if !seen["havoc"] {
				results = append(results, DetectionResult{
					Name: "havoc", Type: provider.TypeHavoc,
					Running: true, Confidence: 0.7,
				})
				seen["havoc"] = true
			}
		}
	}
	return results
}

// checkProcesses looks for C2 server processes.
// Returns an empty (non-nil) slice when the process listing cannot be read —
// a failed ps means "no processes seen", not a failure.
func (r *Runner) checkProcesses(ctx context.Context) []DetectionResult {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ps", "aux")
	out, err := cmd.Output()
	if err != nil {
		return make([]DetectionResult, 0)
	}

	results := make([]DetectionResult, 0)
	output := string(out)

	if strings.Contains(output, "mythic_server") || strings.Contains(output, "mythic-cli") {
		results = append(results, DetectionResult{
			Name: "mythic", Type: provider.TypeMythic,
			Running: true, Confidence: 0.85,
		})
	}
	if strings.Contains(output, "sliver-server") {
		results = append(results, DetectionResult{
			Name: "sliver", Type: provider.TypeSliver,
			Running: true, Confidence: 0.85,
		})
	}
	if strings.Contains(output, "havoc-server") || strings.Contains(output, "havocs-server") {
		results = append(results, DetectionResult{
			Name: "havoc", Type: provider.TypeHavoc,
			Running: true, Confidence: 0.85,
		})
	}
	return results
}

// checkExecProviders scans ~/.c4/providers/*.yaml for existing exec configs.
// Returns an empty (non-nil) slice when the providers directory does not exist
// or holds no YAML configs — a missing directory is "nothing found", not a failure.
func (r *Runner) checkExecProviders() []DetectionResult {
	provDir := filepath.Join(r.homeDir, ".c4", "providers")
	entries, err := os.ReadDir(provDir)
	if err != nil {
		return make([]DetectionResult, 0)
	}

	results := make([]DetectionResult, 0)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".yaml")
		results = append(results, DetectionResult{
			Name: name, Type: provider.TypeExec,
			Path: filepath.Join(provDir, e.Name()),
			Running: false, Confidence: 1.0,
		})
	}
	return results
}

// hasFile checks if a file exists in a directory.
func hasFile(dir, filename string) bool {
	info, err := os.Stat(filepath.Join(dir, filename))
	return err == nil && !info.IsDir()
}
