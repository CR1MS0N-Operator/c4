package detect

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/CR1MS0N-Operator/c4/pkg/provider"
)

func TestCheckInstallPaths_EmptyToolsDir(t *testing.T) {
	r := &Runner{homeDir: t.TempDir()}
	results := r.checkInstallPaths()
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
	if results == nil {
		t.Error("checkInstallPaths() returned nil, want empty slice")
	}
}

func TestCheckInstallPaths_MythicDetected(t *testing.T) {
	home := t.TempDir()
	toolsDir := filepath.Join(home, "Tools", "c2", "mythic")
	if err := os.MkdirAll(toolsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	r := &Runner{homeDir: home}
	results := r.checkInstallPaths()

	found := false
	for _, res := range results {
		if res.Name == "mythic" && res.Type == provider.TypeMythic {
			found = true
			if res.Confidence != 0.9 {
				t.Errorf("Confidence = %f, want 0.9", res.Confidence)
			}
		}
	}
	if !found {
		t.Fatal("mythic not detected in install paths")
	}
}

func TestCheckInstallPaths_MultipleC2s(t *testing.T) {
	home := t.TempDir()
	for _, name := range []string{"mythic", "sliver", "havoc"} {
		dir := filepath.Join(home, "Tools", "c2", name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	r := &Runner{homeDir: home}
	results := r.checkInstallPaths()

	names := make(map[string]bool)
	for _, res := range results {
		names[res.Name] = true
	}

	for _, expected := range []string{"mythic", "sliver", "havoc"} {
		if !names[expected] {
			t.Errorf("%s not detected", expected)
		}
	}
}

func TestCheckInstallPaths_UnknownC2WithComposeFile(t *testing.T) {
	home := t.TempDir()
	unknownDir := filepath.Join(home, "Tools", "c2", "custom-c2")
	if err := os.MkdirAll(unknownDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(unknownDir, "docker-compose.yml"), []byte("version: '3'"), 0o600); err != nil {
		t.Fatal(err)
	}

	r := &Runner{homeDir: home}
	results := r.checkInstallPaths()

	found := false
	for _, res := range results {
		if res.Name == "custom-c2" && res.Type == provider.TypeExec {
			found = true
			if res.Confidence != 0.5 {
				t.Errorf("Confidence = %f, want 0.5", res.Confidence)
			}
		}
	}
	if !found {
		t.Fatal("custom-c2 not detected as exec type")
	}
}

func TestCheckExecProviders_EmptyDir(t *testing.T) {
	r := &Runner{homeDir: t.TempDir()}
	results := r.checkExecProviders()
	if len(results) != 0 {
		t.Errorf("got %d results, want 0", len(results))
	}
	if results == nil {
		t.Error("checkExecProviders() returned nil, want empty slice")
	}
}

func TestCheckExecProviders_WithConfigs(t *testing.T) {
	home := t.TempDir()
	provDir := filepath.Join(home, ".c4", "providers")
	if err := os.MkdirAll(provDir, 0o755); err != nil {
		t.Fatal(err)
	}

	yamlPath := filepath.Join(provDir, "test-c2.yaml")
	if err := os.WriteFile(yamlPath, []byte(`name: test-c2`), 0o600); err != nil {
		t.Fatal(err)
	}

	r := &Runner{homeDir: home}
	results := r.checkExecProviders()

	found := false
	for _, res := range results {
		if res.Name == "test-c2" && res.Type == provider.TypeExec {
			found = true
			if res.Confidence != 1.0 {
				t.Errorf("Confidence = %f, want 1.0", res.Confidence)
			}
		}
	}
	if !found {
		t.Fatal("test-c2 not detected in exec providers")
	}
}

func TestCheckExecProviders_SkipsNonYaml(t *testing.T) {
	home := t.TempDir()
	provDir := filepath.Join(home, ".c4", "providers")
	if err := os.MkdirAll(provDir, 0o755); err != nil {
		t.Fatal(err)
	}

	txtPath := filepath.Join(provDir, "readme.txt")
	if err := os.WriteFile(txtPath, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	r := &Runner{homeDir: home}
	results := r.checkExecProviders()
	if len(results) != 0 {
		t.Errorf("got %d results for non-YAML files, want 0", len(results))
	}
}

func TestRunAll_EmptySystem(t *testing.T) {
	r := &Runner{homeDir: t.TempDir()}
	results := r.RunAll(context.Background())
	if results == nil {
		t.Fatal("RunAll() returned nil, want empty slice")
	}
}

// TestRunAll_EmptySystemMarshalsToArray pins the reason the empty slice
// matters: `c4 detect --json` must emit [] and not null.
func TestRunAll_EmptySystemMarshalsToArray(t *testing.T) {
	r := &Runner{homeDir: t.TempDir()}
	out, err := json.Marshal(r.RunAll(context.Background()))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != "[]" {
		t.Errorf("RunAll() marshalled to %s, want []", out)
	}
}

// TestSubChecksNeverReturnNil locks the package contract on the error paths:
// a sub-check that cannot run reports zero results, not nil.
func TestSubChecksNeverReturnNil(t *testing.T) {
	r := &Runner{homeDir: t.TempDir()}
	checks := map[string][]DetectionResult{
		"checkInstallPaths":  r.checkInstallPaths(),
		"checkDocker":        r.checkDocker(context.Background()),
		"checkProcesses":     r.checkProcesses(context.Background()),
		"checkExecProviders": r.checkExecProviders(),
	}
	for name, results := range checks {
		if results == nil {
			t.Errorf("%s() returned nil, want empty slice", name)
		}
	}
}
