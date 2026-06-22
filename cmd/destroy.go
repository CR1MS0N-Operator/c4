package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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

		switch c2 {
		case "mythic":
			return destroyMythic()
		default:
			return destroyExec(c2)
		}
	},
}

func destroyMythic() error {
	return fmt.Errorf("mythic destroy: not yet implemented (use 'c4 stop mythic')")
}

func destroyExec(name string) error {
	providers, err := loadExecProviders()
	if err != nil {
		return fmt.Errorf("load exec providers: %w", err)
	}

	var target interface{ Destroy(ctx context.Context) error }
	for _, p := range providers {
		if p.Name() == name {
			target = p
			break
		}
	}

	if target == nil {
		return fmt.Errorf("unknown C2: %q (no exec provider found in ~/.c4/providers/)", name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := target.Destroy(ctx); err != nil {
		return fmt.Errorf("destroy %s: %w", name, err)
	}

	fmt.Printf("Exec provider '%s' destroyed\n", name)
	return nil
}

func init() {
	destroyCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
}
