// Package provider defines the C2 provider abstraction and shared types.
package provider

import "time"

// Type represents the kind of C2 provider.
type Type string

// Provider type constants.
const (
	TypeMythic  Type = "Mythic"
	TypeSliver  Type = "Sliver"
	TypeHavoc   Type = "Havoc"
	TypeExec    Type = "Exec"
	TypeUnknown Type = "Unknown"
)

// Status represents the lifecycle state of a C2 instance.
type Status string

// Status constants.
const (
	StatusUnknown   Status = "Unknown"
	StatusDeploying Status = "Deploying"
	StatusRunning   Status = "Running"
	StatusStopped   Status = "Stopped"
	StatusFailed    Status = "Failed"
	StatusDestroyed Status = "Destroyed"
)

// HealthCheckResult represents the outcome of a provider health check.
type HealthCheckResult struct {
	Healthy   bool      `json:"healthy"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}
// Listener represents a C2 listener.
type Listener struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Status   Status `json:"status"`
	Provider Type   `json:"provider"`
}

// Callback represents an active agent callback.
type Callback struct {
	ID       string `json:"id"`
	Listener string `json:"listener"`
	Agent    string `json:"agent"`
	Host     string `json:"host"`
	User     string `json:"user"`
	PID      int    `json:"pid"`
	Status   Status `json:"status"`
}

// Payload represents a generated payload.
type Payload struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Format   string `json:"format"`
	FilePath string `json:"file_path,omitempty"`
}
