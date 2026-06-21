package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ForeverLX/c4/pkg/docker"
	"github.com/spf13/cobra"
)

// startCmd starts an existing C2 instance.
var startCmd = &cobra.Command{
	Use:   "start <c2>",
	Short: "Start an existing C2 instance",
	Long:  `Start a previously deployed or stopped C2 instance via Docker Compose.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c2 := args[0]
		switch c2 {
		case "mythic":
			return startMythic()
		default:
			return fmt.Errorf("unknown C2: %q (supported: mythic)", c2)
		}
	},
}

func startMythic() error {
	if rootConfig == nil {
		return fmt.Errorf("config not loaded")
	}

	localPath := rootConfig.Mythic.LocalPath
	if localPath == "" {
		return fmt.Errorf("mythic.local_path not set in config")
	}

	rootLogger.Info().Str("path", localPath).Msg("starting Mythic")

	dm := docker.NewComposeManager(rootConfig.Docker.Socket, "")
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if err := dm.UpDir(ctx, localPath); err != nil {
		return fmt.Errorf("starting Mythic: %w", err)
	}

	fmt.Printf("Mythic started\n")
	fmt.Printf("  Web UI: https://127.0.0.1:7443\n")
	fmt.Printf("  GraphQL: https://127.0.0.1:7443/graphql/\n")
	return nil
}

func init() {
}
