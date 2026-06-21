package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/CR1MS0N-Operator/c4/pkg/mythic"
	"github.com/CR1MS0N-Operator/c4/pkg/provider"
	"github.com/spf13/cobra"
)

// listenerCmd manages C2 listeners.
var listenerCmd = &cobra.Command{
	Use:   "listener",
	Short: "Manage C2 listeners",
	Long:  `List, start, and stop C2 listeners (C2 profiles in Mythic).`,
}

func getMythicProvider() (*mythic.MythicProvider, error) {
	if rootConfig == nil {
		return nil, fmt.Errorf("config not loaded")
	}
	cfg := rootConfig.Mythic
	return mythic.NewMythicProvider("mythic", cfg), nil
}

func connectMythicProvider(ctx context.Context, p *mythic.MythicProvider) error {
	if rootConfig.Mythic.HasuraSecret == "" {
		return fmt.Errorf("hasura_secret not configured — set in c4.toml or C4_MYTHIC_HASURA_SECRET env var")
	}
	return p.Connect(ctx)
}

func printListenerTable(listeners []provider.Listener) {
	fmt.Printf("%-4s %-24s %-12s %-10s %-6s %-10s\n", "ID", "NAME", "TYPE", "STATUS", "PORT", "PROVIDER")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────")
	for _, l := range listeners {
		fmt.Printf("%-4s %-24s %-12s %-10s %-6d %-10s\n", l.ID, l.Name, l.Type, l.Status, l.Port, l.Provider)
	}
}

// listenerListCmd lists listeners.
var listenerListCmd = &cobra.Command{
	Use:   "list [c2]",
	Short: "List listeners",
	Long:  `List C2 profiles from a provider. Defaults to mythic.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetC2 := "mythic"
		if len(args) == 1 {
			targetC2 = args[0]
		}

		switch targetC2 {
		case "mythic":
			return listMythicListeners()
		default:
			return fmt.Errorf("unknown C2: %q (supported: mythic)", targetC2)
		}
	},
}

func listMythicListeners() error {
	p, err := getMythicProvider()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := connectMythicProvider(ctx, p); err != nil {
		return err
	}
	defer p.Disconnect(context.Background())

	listeners, err := p.Listeners(ctx)
	if err != nil {
		return fmt.Errorf("failed to list listeners: %w", err)
	}

	if len(listeners) == 0 {
		fmt.Println("No C2 profiles found.")
		return nil
	}

	printListenerTable(listeners)
	return nil
}

// listenerStartCmd starts a listener.
var listenerStartCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start a listener",
	Long:  `Start a C2 profile by name. Creates it if it doesn't exist.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		targetC2, _ := cmd.Flags().GetString("c2")
		if targetC2 == "" {
			targetC2 = "mythic"
		}

		switch targetC2 {
		case "mythic":
			return startMythicListener(name)
		default:
			return fmt.Errorf("unknown C2: %q (supported: mythic)", targetC2)
		}
	},
}

func startMythicListener(name string) error {
	p, err := getMythicProvider()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := connectMythicProvider(ctx, p); err != nil {
		return err
	}
	defer p.Disconnect(context.Background())

	result, err := p.StartListener(ctx, provider.Listener{Name: name})
	if err != nil {
		return fmt.Errorf("failed to start listener %q: %w", name, err)
	}

	fmt.Printf("Listener %q (ID: %s) started\n", result.Name, result.ID)
	return nil
}

// listenerStopCmd stops a listener.
var listenerStopCmd = &cobra.Command{
	Use:   "stop <id>",
	Short: "Stop a listener",
	Long:  `Stop a C2 profile by numeric ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		targetC2, _ := cmd.Flags().GetString("c2")
		if targetC2 == "" {
			targetC2 = "mythic"
		}

		switch targetC2 {
		case "mythic":
			return stopMythicListener(id)
		default:
			return fmt.Errorf("unknown C2: %q (supported: mythic)", targetC2)
		}
	},
}

func stopMythicListener(id string) error {
	p, err := getMythicProvider()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := connectMythicProvider(ctx, p); err != nil {
		return err
	}
	defer p.Disconnect(context.Background())

	if err := p.StopListener(ctx, id); err != nil {
		return fmt.Errorf("failed to stop listener %s: %w", id, err)
	}

	fmt.Printf("Listener %s stopped\n", id)
	return nil
}

func init() {
	listenerStartCmd.Flags().String("c2", "mythic", "C2 provider to use")
	listenerStopCmd.Flags().String("c2", "mythic", "C2 provider to use")
	listenerCmd.AddCommand(listenerListCmd)
	listenerCmd.AddCommand(listenerStartCmd)
	listenerCmd.AddCommand(listenerStopCmd)
}
