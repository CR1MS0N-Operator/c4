package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/CR1MS0N-Operator/c4/pkg/config"
	"github.com/CR1MS0N-Operator/c4/pkg/detect"
	"github.com/spf13/cobra"
)

// initCmd creates the initial C4 configuration directory and file.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize C4 configuration",
	Long:  `Create the ~/.c4 directory and a default c4.toml configuration file. Optionally runs detect to find installed C2s.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("unable to determine home directory: %w", err)
		}
		cfgPath := filepath.Join(home, ".c4", "c4.toml")
		if cfgFile != "" {
			cfgPath = cfgFile
		}

		if err := config.Init(cfgPath); err != nil {
			return err
		}

		fmt.Printf("C4 initialized at %s\n", cfgPath)

		// Optionally run detect to find installed C2s
		skipDetect, _ := cmd.Flags().GetBool("skip-detect")
		if !skipDetect {
			fmt.Println()
			fmt.Println("Scanning for installed C2 frameworks...")
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			r := detect.New()
			results := r.RunAll(ctx)

			if len(results) == 0 {
				fmt.Println("  No C2 frameworks detected.")
				fmt.Println("  Add exec providers manually in ~/.c4/providers/ or run 'c4 detect' later.")
			} else {
				fmt.Printf("  Detected %d C2 framework(s):\n", len(results))
				for _, res := range results {
					running := ""
					if res.Running {
						running = " (running)"
					}
					fmt.Printf("    - %s [%s]%s\n", res.Name, res.Type, running)
				}
				fmt.Println("  Run 'c4 detect --apply' to auto-generate provider configs.")
			}
		}

		return nil
	},
}

func init() {
	initCmd.Flags().Bool("skip-detect", false, "skip auto-detection of installed C2s")
}
