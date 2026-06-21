// Package docker shells out to the docker compose CLI to manage C2 deployments.
package docker

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ComposeManager shells out to the docker compose CLI for lifecycle operations.
type ComposeManager struct {
	socket     string
	composeDir string
}

// NewComposeManager returns a new ComposeManager.
func NewComposeManager(socket string, composeDir string) *ComposeManager {
	return &ComposeManager{
		socket:     socket,
		composeDir: composeDir,
	}
}

func (m *ComposeManager) env() []string {
	env := os.Environ()
	if m.socket != "" {
		env = append(env, "DOCKER_HOST="+m.socket)
	}
	return env
}

func (m *ComposeManager) run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", append([]string{"compose"}, args...)...)
	cmd.Env = m.env()
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("docker compose %v failed: %w\n%s", args, err, out.String())
	}
	return out.String(), nil
}

func (m *ComposeManager) runInDir(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", append([]string{"compose"}, args...)...)
	cmd.Env = m.env()
	if dir != "" {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("docker compose %v failed: %w\n%s", args, err, out.String())
	}
	return out.String(), nil
}

// composeFilePath resolves a local_path to a compose file path.
// If localPath is a directory, appends docker-compose.yml.
func composeFilePath(localPath string) string {
	info, err := os.Stat(localPath)
	if err == nil && info.IsDir() {
		return filepath.Join(localPath, "docker-compose.yml")
	}
	return localPath
}

// Up runs docker compose up -d for the named project and compose file.
func (m *ComposeManager) Up(ctx context.Context, name, composeFile string) error {
	file := composeFilePath(composeFile)
	_, err := m.run(ctx, "-p", name, "-f", file, "up", "-d")
	return err
}

// UpDir runs docker compose up -d in the given directory (uses project name from .env).
func (m *ComposeManager) UpDir(ctx context.Context, dir string) error {
	_, err := m.runInDir(ctx, dir, "up", "-d")
	return err
}

// Down runs docker compose down for the named project.
func (m *ComposeManager) Down(ctx context.Context, name string) error {
	_, err := m.run(ctx, "-p", name, "down")
	return err
}

// DownDir runs docker compose down in the given directory.
func (m *ComposeManager) DownDir(ctx context.Context, dir string) error {
	_, err := m.runInDir(ctx, dir, "down")
	return err
}

// Ps returns the process status for the named project.
func (m *ComposeManager) Ps(ctx context.Context, name string) (string, error) {
	return m.run(ctx, "-p", name, "ps")
}

// PsDir returns the process status for the project in the given directory.
func (m *ComposeManager) PsDir(ctx context.Context, dir string) (string, error) {
	return m.runInDir(ctx, dir, "ps")
}

// Logs returns the logs for the named project.
func (m *ComposeManager) Logs(ctx context.Context, name string) (string, error) {
	return m.run(ctx, "-p", name, "logs")
}
