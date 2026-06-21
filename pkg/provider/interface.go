package provider

import "context"

// Provider defines the lifecycle and operational interface for C2 backends.
type Provider interface {
	Type() Type
	Name() string
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Deploy(ctx context.Context) error
	Destroy(ctx context.Context) error
	Health(ctx context.Context) (*HealthCheckResult, error)
	Listeners(ctx context.Context) ([]Listener, error)
	StartListener(ctx context.Context, l Listener) (*Listener, error)
	StopListener(ctx context.Context, id string) error
	Callbacks(ctx context.Context) ([]Callback, error)
	Payloads(ctx context.Context) ([]Payload, error)
	GeneratePayload(ctx context.Context, config map[string]any) (*Payload, error)
}
