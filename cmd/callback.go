package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/CR1MS0N-Operator/c4/pkg/mythic"
	"github.com/CR1MS0N-Operator/c4/pkg/provider"
	"github.com/spf13/cobra"
)

// callbackCmd manages active callbacks.
var callbackCmd = &cobra.Command{
	Use:   "callback",
	Short: "Manage C2 callbacks",
	Long:  `List active callbacks from C2 providers.`,
}

func printCallbackTable(callbacks []provider.Callback) {
	fmt.Printf("%-6s %-20s %-16s %-12s %-6s %-12s\n", "ID", "AGENT", "HOST", "USER", "PID", "STATUS")
	fmt.Println("────────────────────────────────────────────────────────────────────────────")
	for _, c := range callbacks {
		fmt.Printf("%-6s %-20s %-16s %-12s %-6d %-12s\n", c.ID, c.Agent, c.Host, c.User, c.PID, c.Status)
	}
}

// callbackListCmd lists callbacks.
var callbackListCmd = &cobra.Command{
	Use:   "list [c2]",
	Short: "List callbacks",
	Long:  `List active callbacks from a provider. Defaults to mythic.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetC2 := "mythic"
		if len(args) == 1 {
			targetC2 = args[0]
		}
		switch targetC2 {
		case "mythic":
			return listMythicCallbacks()
		default:
			return fmt.Errorf("unknown C2: %q (supported: mythic)", targetC2)
		}
	},
}

func listMythicCallbacks() error {
	if rootConfig == nil {
		return fmt.Errorf("config not loaded")
	}
	if rootConfig.Mythic.HasuraSecret == "" {
		return fmt.Errorf("hasura_secret not configured — set in c4.toml or C4_MYTHIC_HASURA_SECRET env var")
	}

	p := mythic.NewMythicProvider("mythic", rootConfig.Mythic)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := p.Connect(ctx); err != nil {
		return err
	}
	defer p.Disconnect(context.Background())

	callbacks, err := p.Callbacks(ctx)
	if err != nil {
		return fmt.Errorf("failed to list callbacks: %w", err)
	}

	if len(callbacks) == 0 {
		fmt.Println("No active callbacks.")
		return nil
	}

	printCallbackTable(callbacks)
	return nil
}

func init() {
	callbackCmd.AddCommand(callbackListCmd)
}
