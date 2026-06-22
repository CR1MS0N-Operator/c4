// Package cmd implements the Cobra command tree for the C4 CLI.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/CR1MS0N-Operator/c4/pkg/c4lib"
	"github.com/CR1MS0N-Operator/c4/pkg/config"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
	jsonLog bool

	// rootLogger is initialized in PersistentPreRunE and made available to subcommands.
	rootLogger zerolog.Logger

	// rootConfig is the loaded C4 configuration.
	rootConfig *config.Config
)

// rootCmd is the base command for the C4 CLI.
var rootCmd = &cobra.Command{
	Use:     "c4",
	Short:   "C4 — C2 Control Center",
	Long:    `C4 is a command-line control center for managing C2 infrastructure such as Mythic and Sliver.`,
	Version: c4lib.Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level := zerolog.InfoLevel
		if verbose {
			level = zerolog.DebugLevel
		}
		zerolog.SetGlobalLevel(level)
		if jsonLog {
			rootLogger = zerolog.New(os.Stderr).With().Timestamp().Logger()
		} else {
			rootLogger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: false}).With().Timestamp().Logger()
		}

		if cfgFile == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("unable to determine home directory: %w", err)
			}
			cfgFile = filepath.Join(home, ".c4", "c4.toml")
		}

		if _, err := os.Stat(cfgFile); err == nil {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load config %q: %w", cfgFile, err)
			}
			rootConfig = cfg
		} else if os.IsNotExist(err) {
			rootConfig = config.DefaultConfig()
		} else {
			return fmt.Errorf("failed to stat config %q: %w", cfgFile, err)
		}

		viper.SetConfigFile(cfgFile)
		return nil
	},
}

// Execute adds all child commands to the root command and executes it.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.c4/c4.toml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose (debug) logging")
	rootCmd.PersistentFlags().BoolVar(&jsonLog, "json", false, "output logs as JSON")
}

func addSubcommands() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(destroyCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(listenerCmd)
	rootCmd.AddCommand(callbackCmd)
	rootCmd.AddCommand(payloadCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(detectCmd)
}

func init() {
	addSubcommands()
}
