package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/CR1MS0N-Operator/c4/pkg/execprovider"
	"github.com/CR1MS0N-Operator/c4/pkg/mythic"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

func printHealthTable(log zerolog.Logger, name, c2Type, status, host, version, message string) {
	fmt.Printf("%-16s %-10s %-12s %-22s %-10s\n", name, c2Type, status, host, version)
	if message != "" {
		fmt.Printf("  └─ %s\n", message)
	}
}

// loadExecProviders scans ~/.c4/providers/*.yaml and returns a list of providers.
func loadExecProviders() ([]*execprovider.Provider, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := home + "/.c4/providers"
	configs, err := execprovider.LoadDir(dir)
	if err != nil {
		return nil, err
	}
	var providers []*execprovider.Provider
	for _, cfg := range configs {
		providers = append(providers, execprovider.New(cfg))
	}
	return providers, nil
}

// statusCmd shows the status of C2 instances.
var statusCmd = &cobra.Command{
	Use:   "status [c2]",
	Short: "Show C2 instance status",
	Long:  `Display the status of all configured C2 instances, or a specific one if provided.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if rootConfig == nil {
			return fmt.Errorf("config not loaded")
		}

		targetC2 := ""
		if len(args) == 1 {
			targetC2 = args[0]
		}

		rootLogger.Debug().Str("config_host", rootConfig.Mythic.Host).Int("config_port", rootConfig.Mythic.APIPort).Bool("config_ssl", rootConfig.Mythic.SSL).Msg("status check")

		fmt.Printf("%-16s %-10s %-12s %-22s %-10s\n", "NAME", "TYPE", "STATUS", "HOST", "VERSION")
		fmt.Println(string(rune(0x2500)) + "────────────────────────────────────────────────────────────────────────────")

		// Mythic
		if targetC2 == "" || targetC2 == "mythic" {
			displayMythicStatus()
		}

		// Exec providers from ~/.c4/providers/*.yaml
		execProviders, err := loadExecProviders()
		if err != nil {
			rootLogger.Warn().Err(err).Msg("failed to load exec providers")
		}
		for _, ep := range execProviders {
			if targetC2 != "" && targetC2 != ep.Name() {
				continue
			}
			displayExecStatus(ep)
		}

		if targetC2 == "sliver" {
			fmt.Printf("%-16s %-10s %-12s %-22s %-10s\n", "sliver", "Sliver", "Not configured", "", "")
		}
		if targetC2 == "havoc" {
			fmt.Printf("%-16s %-10s %-12s %-22s %-10s\n", "havoc", "Havoc", "Not configured", "", "")
		}

		return nil
	},
}

func displayMythicStatus() {
	cfg := rootConfig.Mythic
	p := mythic.NewMythicProvider("mythic", cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	host := fmt.Sprintf("%s:%d", cfg.Host, cfg.APIPort)
	version := cfg.Version
	if version == "" {
		version = "latest"
	}

	if cfg.HasuraSecret == "" {
		printHealthTable(rootLogger, "mythic", "Mythic", "No secret", host, version, "hasura_secret not configured")
		return
	}

	if err := p.Connect(ctx); err != nil {
		printHealthTable(rootLogger, "mythic", "Mythic", "Unreachable", host, version, err.Error())
		return
	}
	defer p.Disconnect(context.Background())

	health, err := p.Health(ctx)
	if err != nil {
		printHealthTable(rootLogger, "mythic", "Mythic", "Error", host, version, err.Error())
		return
	}

	if health.Healthy {
		printHealthTable(rootLogger, "mythic", "Mythic", "Running", host, version, health.Message)
	} else {
		printHealthTable(rootLogger, "mythic", "Mythic", "Unhealthy", host, version, health.Message)
	}
}

func displayExecStatus(p *execprovider.Provider) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	name := p.Name()

	if err := p.Connect(ctx); err != nil {
		printHealthTable(rootLogger, name, "Exec", "No config", "", "", err.Error())
		return
	}

	health, err := p.Health(ctx)
	if err != nil {
		printHealthTable(rootLogger, name, "Exec", "Error", "", "", err.Error())
		return
	}

	if health.Healthy {
		printHealthTable(rootLogger, name, "Exec", "Running", "", "", health.Message)
	} else {
		printHealthTable(rootLogger, name, "Exec", "Stopped", "", "", health.Message)
	}
}
