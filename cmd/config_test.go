package cmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CR1MS0N-Operator/c4/pkg/config"
)

// runConfigSet runs `c4 config set` against cfg with stdin as standard input
// and returns what it printed to stdout.
func runConfigSet(t *testing.T, cfg, stdin string, args ...string) (string, error) {
	t.Helper()

	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := inW.WriteString(stdin); err != nil {
		t.Fatal(err)
	}
	inW.Close()
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	oldIn, oldOut, oldCfg := os.Stdin, os.Stdout, cfgFile
	os.Stdin, os.Stdout, cfgFile = inR, outW, cfg
	defer func() { os.Stdin, os.Stdout, cfgFile = oldIn, oldOut, oldCfg }()

	runErr := configSetCmd.RunE(configSetCmd, args)
	outW.Close()
	out, err := io.ReadAll(outR)
	if err != nil {
		t.Fatal(err)
	}
	return string(out), runErr
}

func TestConfigSetSecretFromStdin(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".c4", "c4.toml")

	for _, args := range [][]string{{"mythic.hasura_secret"}, {"mythic.hasura_secret", "-"}} {
		out, err := runConfigSet(t, path, "s3cret\n", args...)
		if err != nil {
			t.Fatalf("config set %v: %v", args, err)
		}
		if strings.Contains(out, "s3cret") {
			t.Errorf("config set %v echoed the secret: %q", args, out)
		}
		cfg, err := config.Load(path)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Mythic.HasuraSecret != "s3cret" {
			t.Errorf("hasura_secret = %q", cfg.Mythic.HasuraSecret)
		}
		fi, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if got := fi.Mode().Perm(); got != 0o600 {
			t.Errorf("file mode = %o, want 600", got)
		}
		os.Remove(path)
	}
}

func TestConfigSetSecretRejectedOnCommandLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c4.toml")

	for _, key := range []string{"mythic.hasura_secret", "Mythic.Hasura_Secret"} {
		out, err := runConfigSet(t, path, "", key, "s3cret")
		if err == nil {
			t.Fatalf("config set %s accepted a secret on the command line", key)
		}
		if strings.Contains(out, "s3cret") || strings.Contains(err.Error(), "s3cret") {
			t.Errorf("secret leaked in output: %q / %v", out, err)
		}
		if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
			t.Errorf("config file was written: %v", statErr)
		}
	}
}

func TestConfigSetKeepsOtherKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c4.toml")
	if err := os.WriteFile(path, []byte("[mythic]\nhasura_secret = \"keep\"\napi_port = 9443\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := runConfigSet(t, path, "", "mythic.host", "10.0.0.5")
	if err != nil {
		t.Fatalf("config set: %v", err)
	}
	if !strings.Contains(out, "Set mythic.host = 10.0.0.5") {
		t.Errorf("unexpected output: %q", out)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mythic.Host != "10.0.0.5" || cfg.Mythic.HasuraSecret != "keep" || cfg.Mythic.APIPort != 9443 {
		t.Errorf("config after set: %+v", cfg.Mythic)
	}
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fi.Mode().Perm(); got != 0o600 {
		t.Errorf("file mode = %o, want 600", got)
	}
}

func TestConfigSetEmptyStdin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c4.toml")
	if _, err := runConfigSet(t, path, "", "mythic.hasura_secret"); err == nil {
		t.Fatal("config set accepted an empty value")
	}
}
