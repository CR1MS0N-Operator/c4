package execprovider

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	if err := os.WriteFile(path, []byte(`
type: exec
name: "my-c2"
start: "echo start"
stop: "echo stop"
health:
  cmd: "echo ok"
timeout: 15
`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Name != "my-c2" {
		t.Errorf("Name = %q, want %q", cfg.Name, "my-c2")
	}
	if cfg.StartCmd != "echo start" {
		t.Errorf("StartCmd = %q", cfg.StartCmd)
	}
	if cfg.Timeout != 15 {
		t.Errorf("Timeout = %d, want 15", cfg.Timeout)
	}
}

func TestLoad_MissingName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	if err := os.WriteFile(path, []byte(`type: exec`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestLoad_WrongType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	if err := os.WriteFile(path, []byte(`
type: mythic
name: "oops"
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for wrong type, got nil")
	}
}

func TestLoad_DefaultTimeout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	if err := os.WriteFile(path, []byte(`
type: exec
name: "defaults"
start: "echo"
`), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Timeout != 30 {
		t.Errorf("Timeout = %d, want 30", cfg.Timeout)
	}
}

func TestLoadDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	configs, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir() error = %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("got %d configs, want 0", len(configs))
	}
}

func TestLoadDir_PartialFailure(t *testing.T) {
	dir := t.TempDir()

	// Valid YAML
	valid := filepath.Join(dir, "good.yaml")
	if err := os.WriteFile(valid, []byte(`
type: exec
name: "good"
start: "echo good"
`), 0o600); err != nil {
		t.Fatal(err)
	}

	// Invalid YAML (wrong type)
	invalid := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(invalid, []byte(`type: invalid`), 0o600); err != nil {
		t.Fatal(err)
	}

	// LoadDir should return partial results + error
	configs, err := LoadDir(dir)
	if err == nil {
		t.Fatal("expected error for partial failure, got nil")
	}
	if len(configs) != 1 {
		t.Errorf("got %d configs, want 1 (good.yaml)", len(configs))
	}
}

func TestConnect_MissingBinary(t *testing.T) {
	cfg := ExecConfig{
		Name:     "test",
		StartCmd: "./nonexistent-binary-xyz --flag",
		Timeout:  10,
	}
	p := New(cfg)
	if err := p.Connect(nil); err == nil {
		t.Fatal("expected error for missing binary, got nil")
	}
}

func TestConnect_EmptyStartCmd(t *testing.T) {
	cfg := ExecConfig{Name: "test", StartCmd: ""}
	p := New(cfg)
	if err := p.Connect(nil); err == nil {
		t.Fatal("expected error for empty start command, got nil")
	}
}

func TestProvider_NameType(t *testing.T) {
	cfg := ExecConfig{Name: "my-c2"}
	p := New(cfg)
	if p.Name() != "my-c2" {
		t.Errorf("Name() = %q", p.Name())
	}
	if p.Type() != "Exec" {
		t.Errorf("Type() = %q", p.Type())
	}
}

func TestLoad_NonexistentFile(t *testing.T) {
	if _, err := Load("/nonexistent/path.yaml"); err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestLoadDir_NonexistentDir(t *testing.T) {
	configs, err := LoadDir("/nonexistent/dir")
	if err != nil {
		t.Fatalf("LoadDir() error = %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("got %d configs, want 0", len(configs))
	}
}
