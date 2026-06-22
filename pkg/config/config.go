// Package config manages C4 configuration loading, saving, and initialization.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config is the top-level C4 configuration structure.
type Config struct {
	Defaults DefaultsConfig `mapstructure:"defaults"`
	Mythic   C2Config       `mapstructure:"mythic"`
	Docker   DockerConfig   `mapstructure:"docker"`
}

// DefaultsConfig contains default settings applied across the CLI.
type DefaultsConfig struct {
	DataDir  string `mapstructure:"data_dir"`
	LogLevel string `mapstructure:"log_level"`
}

// C2Config contains connection details for a C2 provider.
type C2Config struct {
	Host         string `mapstructure:"host"`
	APIPort      int    `mapstructure:"api_port"`
	ServerPort   int    `mapstructure:"server_port"`
	HasuraSecret string `mapstructure:"hasura_secret"`
	SSL          bool   `mapstructure:"ssl"`
	LocalPath    string `mapstructure:"local_path"`
	DataDir      string `mapstructure:"data_dir"`
	Version      string `mapstructure:"version"`
}

// DockerConfig contains Docker daemon and compose directory settings.
type DockerConfig struct {
	Socket     string `mapstructure:"socket"`
	ComposeDir string `mapstructure:"compose_dir"`
}

// DefaultConfig returns a Config populated with safe defaults.
func DefaultConfig() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dataDir := filepath.Join(home, ".c4", "data")
	return &Config{
		Defaults: DefaultsConfig{
			DataDir:  dataDir,
			LogLevel: "info",
		},
		Mythic: C2Config{
			Host:         "127.0.0.1",
			APIPort:      7443,
			ServerPort:   17443,
			HasuraSecret: "",
			SSL:          true,
			LocalPath:    filepath.Join(home, "Tools", "c2", "mythic"),
			DataDir:      filepath.Join(dataDir, "mythic"),
			Version:      "latest",
		},
		Docker: DockerConfig{
			Socket:     "/var/run/docker.sock",
			ComposeDir: filepath.Join(dataDir, "compose"),
		},
	}
}

// defaultConfigText is the content written when initializing a new config file.
const defaultConfigText = `# C4 (C2 Control Center) configuration file
# See https://github.com/CR1MS0N-Operator/c4 for documentation.

[defaults]
data_dir = "~/.c4/data"
log_level = "info"

[mythic]
host = "127.0.0.1"
api_port = 7443
server_port = 17443
hasura_secret = ""
ssl = true
local_path = "~/Tools/c2/mythic"
data_dir = "~/.c4/data/mythic"
version = "latest"

[docker]
socket = "/var/run/docker.sock"
compose_dir = "~/.c4/data/compose"
`

// Load reads a TOML config from the given path into a Config value.
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config %q: %w", path, err)
	}

	cfg := DefaultConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return cfg, nil
}

// Save writes the provided Config to the given path as TOML.
func Save(path string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory %q: %w", dir, err)
	}

	v := viper.New()
	v.SetConfigType("toml")
	if err := v.MergeConfigMap(map[string]any{
		"defaults": map[string]any{
			"data_dir":  cfg.Defaults.DataDir,
			"log_level": cfg.Defaults.LogLevel,
		},
		"mythic": map[string]any{
			"host":          cfg.Mythic.Host,
			"api_port":      cfg.Mythic.APIPort,
			"server_port":   cfg.Mythic.ServerPort,
			"hasura_secret": cfg.Mythic.HasuraSecret,
			"ssl":           cfg.Mythic.SSL,
			"local_path":    cfg.Mythic.LocalPath,
			"data_dir":      cfg.Mythic.DataDir,
			"version":       cfg.Mythic.Version,
		},
		"docker": map[string]any{
			"socket":      cfg.Docker.Socket,
			"compose_dir": cfg.Docker.ComposeDir,
		},
	}); err != nil {
		return fmt.Errorf("failed to merge config map: %w", err)
	}

	if err := v.WriteConfigAs(path); err != nil {
		return fmt.Errorf("failed to write config %q: %w", path, err)
	}
	return nil
}

// Init creates a default config file at the given path if it does not exist.
// Also creates the ~/.c4/providers/ directory for exec provider YAML configs.
func Init(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config already exists at %q", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat config path %q: %w", path, err)
	}

	dir := filepath.Dir(path)
	providersDir := filepath.Join(dir, "providers")

	if err := os.MkdirAll(providersDir, 0o755); err != nil {
		return fmt.Errorf("failed to create providers directory %q: %w", providersDir, err)
	}

	if err := os.WriteFile(path, []byte(defaultConfigText), 0o600); err != nil {
		return fmt.Errorf("failed to write config %q: %w", path, err)
	}
	return nil
}
