package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// payloadCmd manages C2 payloads.
var payloadCmd = &cobra.Command{
	Use:   "payload",
	Short: "Manage C2 payloads",
	Long:  `Generate and list C2 payloads.`,
}

// payloadGenerateCmd generates a payload.
var payloadGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a payload",
	RunE: func(cmd *cobra.Command, args []string) error {
		c2, _ := cmd.Flags().GetString("c2")
		fmt.Printf("Generating payload for %s... (not yet implemented)\n", c2)
		return nil
	},
}

// payloadListCmd lists payloads.
var payloadListCmd = &cobra.Command{
	Use:   "list [c2]",
	Short: "List payloads",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c2 := "all"
		if len(args) == 1 {
			c2 = args[0]
		}
		fmt.Printf("Listing payloads for %s... (not yet implemented)\n", c2)
		return nil
	},
}

func init() {
	payloadGenerateCmd.Flags().String("c2", "mythic", "C2 provider to use")
	payloadCmd.AddCommand(payloadGenerateCmd)
	payloadCmd.AddCommand(payloadListCmd)
}
