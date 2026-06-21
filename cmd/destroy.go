package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// destroyCmd tears down a C2 instance.
var destroyCmd = &cobra.Command{
	Use:   "destroy <c2>",
	Short: "Destroy a C2 instance",
	Long:  `Destroy a deployed C2 instance. Use --force to skip confirmation.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c2 := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if !force {
			fmt.Printf("Are you sure you want to destroy the %s instance? [y/N]: ", c2)
			reader := bufio.NewReader(os.Stdin)
			resp, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}
			resp = strings.TrimSpace(strings.ToLower(resp))
			if resp != "y" && resp != "yes" {
				fmt.Println("Destroy cancelled.")
				return nil
			}
		}

		fmt.Printf("Destroying %s instance... (not yet implemented)\n", c2)
		return nil
	},
}

func init() {
	destroyCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
}
