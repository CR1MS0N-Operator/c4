package config

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func mode(t *testing.T, path string) os.FileMode {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	return fi.Mode().Perm()
}

func TestSetCreatesPrivateFileAndDir(t *testing.T) {
	old := syscall.Umask(0o022)
	defer syscall.Umask(old)

	dir := filepath.Join(t.TempDir(), ".c4")
	path := filepath.Join(dir, "c4.toml")

	if err := Set(path, "mythic.hasura_secret", "s3cret"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got := mode(t, path); got != 0o600 {
		t.Errorf("file mode = %o, want 600", got)
	}
	if got := mode(t, dir); got != 0o700 {
		t.Errorf("dir mode = %o, want 700", got)
	}
}

func TestSetTightensExistingFileAndKeepsOtherKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c4.toml")
	existing := "[mythic]\nhost = \"10.0.0.5\"\napi_port = 9443\n\n[docker]\nsocket = \"/run/docker.sock\"\n"
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Set(path, "mythic.hasura_secret", "s3cret"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got := mode(t, path); got != 0o600 {
		t.Errorf("file mode = %o, want 600", got)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Mythic.HasuraSecret != "s3cret" {
		t.Errorf("hasura_secret = %q", cfg.Mythic.HasuraSecret)
	}
	if cfg.Mythic.Host != "10.0.0.5" || cfg.Mythic.APIPort != 9443 || cfg.Docker.Socket != "/run/docker.sock" {
		t.Errorf("existing keys were dropped: %+v", cfg)
	}
}

func TestSetRejectsUnreadableConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c4.toml")
	bad := "[mythic\nhost = "
	if err := os.WriteFile(path, []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Set(path, "mythic.host", "x"); err == nil {
		t.Fatal("Set succeeded on a malformed config")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != bad {
		t.Errorf("malformed config was overwritten: %q", got)
	}
}

func TestSaveWritesPrivateFile(t *testing.T) {
	old := syscall.Umask(0o022)
	defer syscall.Umask(old)

	dir := filepath.Join(t.TempDir(), ".c4")
	path := filepath.Join(dir, "c4.toml")
	cfg := DefaultConfig()
	cfg.Mythic.HasuraSecret = "s3cret"

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got := mode(t, path); got != 0o600 {
		t.Errorf("file mode = %o, want 600", got)
	}
	if got := mode(t, dir); got != 0o700 {
		t.Errorf("dir mode = %o, want 700", got)
	}

	// An existing world-readable file is tightened on rewrite.
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if got := mode(t, path); got != 0o600 {
		t.Errorf("file mode after rewrite = %o, want 600", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "s3cret") {
		t.Error("saved config is missing hasura_secret")
	}
}

func TestInitCreatesPrivateFileAndDirs(t *testing.T) {
	old := syscall.Umask(0o022)
	defer syscall.Umask(old)

	dir := filepath.Join(t.TempDir(), ".c4")
	path := filepath.Join(dir, "c4.toml")
	if err := Init(path); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if got := mode(t, path); got != 0o600 {
		t.Errorf("file mode = %o, want 600", got)
	}
	for _, d := range []string{dir, filepath.Join(dir, "providers")} {
		if got := mode(t, d); got != 0o700 {
			t.Errorf("%s mode = %o, want 700", d, got)
		}
	}
}
