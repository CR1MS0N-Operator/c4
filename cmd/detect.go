package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/CR1MS0N-Operator/c4/pkg/detect"
	"github.com/spf13/cobra"
)

// detectCmd scans for installed C2 frameworks.
var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Auto-detect installed C2 frameworks",
	Long:  `Scan the system for installed C2 frameworks (install paths, Docker containers, processes, existing exec provider configs).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apply, _ := cmd.Flags().GetBool("apply")
		asJSON, _ := cmd.Flags().GetBool("json")
		return runDetect(apply, asJSON)
	},
}

func runDetect(apply, asJSON bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	r := detect.New()
	results := r.RunAll(ctx)

	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	if len(results) == 0 {
		fmt.Println("No C2 frameworks detected.")
		return nil
	}

	fmt.Printf("%-12s %-10s %-36s %-8s %-6s %-10s\n", "NAME", "TYPE", "PATH", "RUNNING", "PORT", "CONFIDENCE")
	fmt.Println("──────────────────────────────────────────────────────────────────────────────────────")
	for _, res := range results {
		path := res.Path
		if path == "" {
			path = "-"
		}
		running := "no"
		if res.Running {
			running = "yes"
		}
		port := "-"
		if res.Port > 0 {
			port = fmt.Sprintf("%d", res.Port)
		}
		fmt.Printf("%-12s %-10s %-36s %-8s %-6s %-10.0f%%\n", res.Name, res.Type, path, running, port, res.Confidence*100)
	}

	if apply {
		return applyDetect(results)
	}
	return nil
}

func applyDetect(results []detect.DetectionResult) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}

	providersDir := filepath.Join(home, ".c4", "providers")
	if err := os.MkdirAll(providersDir, 0o755); err != nil {
		return fmt.Errorf("create providers dir: %w", err)
	}

	applied := 0
	for _, res := range results {
		if res.Type == "Mythic" || res.Type == "Exec" {
			// Skip Mythic — already in c4.toml, not exec config
			// Skip exec — already configured
			continue
		}
		// Auto-generate exec provider YAML for detected C2s
		yamlPath := filepath.Join(providersDir, res.Name+".yaml")
		if _, err := os.Stat(yamlPath); err == nil {
			continue // already exists
		}

		safeName := res.Name
		content := fmt.Sprintf(`type: exec
name: "%s"
start: "echo 'start %s — configure me in %s'"
stop: "echo 'stop %s'"
health:
  cmd: "echo ok"
timeout: 30
`, safeName, safeName, res.Path, safeName)

		if err := os.WriteFile(yamlPath, []byte(content), 0o600); err != nil {
			return fmt.Errorf("write %s: %w", yamlPath, err)
		}
		fmt.Printf("  Created %s\n", yamlPath)
		applied++
	}

	if applied == 0 {
		fmt.Println("No new providers to generate.")
	} else {
		fmt.Printf("Generated %d provider config(s). Edit them to add real start/stop/health commands.\n", applied)
	}
	return nil
}

func init() {
	detectCmd.Flags().Bool("apply", false, "auto-generate c4.toml entries and exec provider YAMLs")
	detectCmd.Flags().Bool("json", false, "output as JSON")
}
