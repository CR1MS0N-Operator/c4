package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ForeverLX/c4/pkg/docker"
	"github.com/spf13/cobra"
)

// deployCmd deploys a C2 instance.
var deployCmd = &cobra.Command{
	Use:   "deploy <c2>",
	Short: "Deploy or start a C2 instance",
	Long:  `Deploy or start a C2 instance such as mythic. For local instances, this starts the existing Docker Compose deployment.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c2 := args[0]
		instanceName, _ := cmd.Flags().GetString("name")
		if instanceName == "" {
			instanceName = c2
		}

		switch c2 {
		case "mythic":
			return deployMythic(instanceName)
		default:
			return fmt.Errorf("unknown C2: %q (supported: mythic)", c2)
		}
	},
}

func deployMythic(name string) error {
	if rootConfig == nil {
		return fmt.Errorf("config not loaded")
	}

	localPath := rootConfig.Mythic.LocalPath
	if localPath == "" {
		return fmt.Errorf("mythic.local_path not set in config")
	}

	rootLogger.Info().Str("name", name).Str("path", localPath).Msg("starting Mythic")

	dm := docker.NewComposeManager(rootConfig.Docker.Socket, "")
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if err := dm.Up(ctx, name, localPath); err != nil {
		return fmt.Errorf("starting Mythic: %w", err)
	}

	fmt.Printf("Mythic '%s' started at %s\n", name, localPath)
	fmt.Printf("  Web UI: https://127.0.0.1:7443\n")
	fmt.Printf("  GraphQL: https://127.0.0.1:7443/graphql/\n")
	return nil
}

func init() {
	deployCmd.Flags().String("name", "", "instance name (defaults to c2 name)")
}
