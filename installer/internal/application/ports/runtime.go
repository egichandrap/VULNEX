package ports

import "context"

// RuntimePort manages container orchestration via Docker Compose.
type RuntimePort interface {
	Pull(ctx context.Context, composePath string, envFile string) error
	Up(ctx context.Context, composePath string, envFile string) error
	Down(ctx context.Context, composePath string, removeVolumes bool) error
	Ps(ctx context.Context, composePath string) (services map[string]string, err error)
	Logs(ctx context.Context, composePath string, service string, follow bool) error
	Exec(ctx context.Context, composePath string, service string, command []string) (string, error)
}
