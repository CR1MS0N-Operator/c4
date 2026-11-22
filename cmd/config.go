package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/CR1MS0N-Operator/c4/pkg/config"
	"github.com/spf13/cobra"
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
	Use:   "set <key> [value|-]",
	Short: "Set a config value",
	Long: `Set a config value in the config file, keeping all other keys.

If value is omitted or "-", it is read from stdin (prompted for, without echo,
on a terminal). Keys containing "secret" must be given this way so the value
never appears in the process list or shell history, and their value is not
printed.`,
	Example: `  c4 config set mythic.host 10.0.0.5
  c4 config set mythic.hasura_secret            # prompts without echo
  printf '%s' "$SECRET" | c4 config set mythic.hasura_secret -`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		secret := isSecretKey(key)

		var value string
		if len(args) == 2 && args[1] != "-" {
			if secret {
				return fmt.Errorf("refusing to take %q from the command line; omit the value to be prompted, or pipe it on stdin with \"-\"", key)
			}
			value = args[1]
		} else {
			v, err := readValue(os.Stdin, key)
			if err != nil {
				return err
			}
			value = v
		}

		if err := config.Set(cfgFile, key, value); err != nil {
			return err
		}
		if secret {
			fmt.Printf("Set %s (value hidden)\n", key)
		} else {
			fmt.Printf("Set %s = %s\n", key, value)
		}
		return nil
	},
}

func isSecretKey(key string) bool {
	return strings.Contains(strings.ToLower(key), "secret")
}

// readValue reads a config value from in, prompting without echo when in is
// a terminal.
func readValue(in *os.File, key string) (string, error) {
	if isTerminal(in) {
		return readHidden(in, fmt.Sprintf("Value for %s: ", key))
	}
	return readLine(bufio.NewReader(in))
}

// readLine reads one line from r without its trailing line ending.
func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read value: %w", err)
	}
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return "", errors.New("no value given")
	}
	return line, nil
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
