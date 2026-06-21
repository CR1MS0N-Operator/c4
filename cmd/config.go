package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/CR1MS0N-Operator/c4/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd manages C4 configuration.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage C4 configuration",
	Long:  `Initialize, view, and modify the C4 configuration file.`,
}

// configInitCmd initializes the config file.
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the C4 config file",
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

// configSetCmd sets a config value.
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]
		viper.Set(key, value)
		if err := viper.WriteConfig(); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

// configViewCmd prints the current config.
var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View the current config",
	RunE: func(cmd *cobra.Command, args []string) error {
		if rootConfig == nil {
			return fmt.Errorf("no configuration loaded")
		}
		fmt.Printf("Defaults:\n  DataDir:  %s\n  LogLevel: %s\n", rootConfig.Defaults.DataDir, rootConfig.Defaults.LogLevel)
		fmt.Printf("Mythic:\n  Host:    %s\n  APIPort: %d\n  SSL:     %t\n  DataDir: %s\n  Version: %s\n",
			rootConfig.Mythic.Host, rootConfig.Mythic.APIPort, rootConfig.Mythic.SSL, rootConfig.Mythic.DataDir, rootConfig.Mythic.Version)
		fmt.Printf("Docker:\n  Socket:     %s\n  ComposeDir: %s\n", rootConfig.Docker.Socket, rootConfig.Docker.ComposeDir)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configViewCmd)
}
