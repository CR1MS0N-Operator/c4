package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// callbackCmd manages active callbacks.
var callbackCmd = &cobra.Command{
	Use:   "callback",
	Short: "Manage C2 callbacks",
	Long:  `List active callbacks from C2 providers.`,
}

// callbackListCmd lists callbacks.
var callbackListCmd = &cobra.Command{
	Use:   "list [c2]",
	Short: "List callbacks",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c2 := "all"
		if len(args) == 1 {
			c2 = args[0]
		}
		fmt.Printf("Listing callbacks for %s... (not yet implemented)\n", c2)
		return nil
	},
}

func init() {
	callbackCmd.AddCommand(callbackListCmd)
}
