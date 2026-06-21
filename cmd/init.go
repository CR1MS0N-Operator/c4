package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ForeverLX/c4/pkg/config"
	"github.com/spf13/cobra"
)

// initCmd creates the initial C4 configuration directory and file.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize C4 configuration",
	Long:  `Create the ~/.c4 directory and a default c4.toml configuration file.`,
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
		return nil
	},
}
