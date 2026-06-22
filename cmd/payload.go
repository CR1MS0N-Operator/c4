package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/CR1MS0N-Operator/c4/pkg/mythic"
	"github.com/CR1MS0N-Operator/c4/pkg/provider"
	"github.com/spf13/cobra"
)

// payloadCmd manages C2 payloads.
var payloadCmd = &cobra.Command{
	Use:   "payload",
	Short: "Manage C2 payloads",
	Long:  `Generate and list C2 payloads.`,
}

func printPayloadTable(payloads []provider.Payload) {
	fmt.Printf("%-6s %-30s %-16s %-12s %-36s\n", "ID", "NAME", "TYPE", "FORMAT", "UUID")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────────────────────")
	for _, p := range payloads {
		fmt.Printf("%-6s %-30s %-16s %-12s %-36s\n", p.ID, p.Name, p.Type, p.Format, p.FilePath)
	}
}

// payloadListCmd lists payloads.
var payloadListCmd = &cobra.Command{
	Use:   "list [c2]",
	Short: "List payloads",
	Long:  `List payloads from a provider. Defaults to mythic.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetC2 := "mythic"
		if len(args) == 1 {
			targetC2 = args[0]
		}
		switch targetC2 {
		case "mythic":
			return listMythicPayloads()
		default:
			return fmt.Errorf("unknown C2: %q (supported: mythic)", targetC2)
		}
	},
}

func listMythicPayloads() error {
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

	payloads, err := p.Payloads(ctx)
	if err != nil {
		return fmt.Errorf("failed to list payloads: %w", err)
	}

	if len(payloads) == 0 {
		fmt.Println("No payloads found.")
		return nil
	}

	printPayloadTable(payloads)
	return nil
}

// payloadGenerateCmd generates a payload.
var payloadGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a payload",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented")
	},
}

func init() {
	payloadGenerateCmd.Flags().String("c2", "mythic", "C2 provider to use")
	payloadCmd.AddCommand(payloadGenerateCmd)
	payloadCmd.AddCommand(payloadListCmd)
}
