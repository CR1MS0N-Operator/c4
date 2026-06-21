package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/CR1MS0N-Operator/c4/pkg/docker"
	"github.com/spf13/cobra"
)

// stopCmd stops a running C2 instance.
var stopCmd = &cobra.Command{
	Use:   "stop <c2>",
	Short: "Stop a running C2 instance",
	Long:  `Stop a C2 instance via Docker Compose down.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c2 := args[0]
		switch c2 {
		case "mythic":
			return stopMythic()
		default:
			return fmt.Errorf("unknown C2: %q (supported: mythic)", c2)
		}
	},
}

func stopMythic() error {
	if rootConfig == nil {
		return fmt.Errorf("config not loaded")
	}

	localPath := rootConfig.Mythic.LocalPath
	if localPath == "" {
		return fmt.Errorf("mythic.local_path not set in config")
	}

	rootLogger.Info().Str("path", localPath).Msg("stopping Mythic")

	dm := docker.NewComposeManager(rootConfig.Docker.Socket, "")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := dm.DownDir(ctx, localPath); err != nil {
		return fmt.Errorf("stopping Mythic: %w", err)
	}

	fmt.Printf("Mythic stopped\n")
	return nil
}

func init() {
}
