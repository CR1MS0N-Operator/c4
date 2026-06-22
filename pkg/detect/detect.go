// Package detect auto-discovers installed C2 frameworks on the local system.
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
func (r *Runner) RunAll(ctx context.Context) []DetectionResult {
	var results []DetectionResult
	results = append(results, r.checkInstallPaths()...)
	results = append(results, r.checkDocker(ctx)...)
	results = append(results, r.checkProcesses(ctx)...)
	results = append(results, r.checkExecProviders()...)
	return results
}

// checkInstallPaths looks for C2 install directories in ~/Tools/c2/.
func (r *Runner) checkInstallPaths() []DetectionResult {
	toolsDir := filepath.Join(r.homeDir, "Tools", "c2")
	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		return nil
	}

	var results []DetectionResult

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
func (r *Runner) checkDocker(ctx context.Context) []DetectionResult {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker", "ps", "--format", "{{.Names}}")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var results []DetectionResult
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
func (r *Runner) checkProcesses(ctx context.Context) []DetectionResult {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ps", "aux")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var results []DetectionResult
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
func (r *Runner) checkExecProviders() []DetectionResult {
	provDir := filepath.Join(r.homeDir, ".c4", "providers")
	entries, err := os.ReadDir(provDir)
	if err != nil {
		return nil
	}

	var results []DetectionResult
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
