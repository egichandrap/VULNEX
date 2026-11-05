package ports

import "context"

// EnvironmentDetector inspects host system capabilities and dependencies.
type EnvironmentDetector interface {
	DetectOS(ctx context.Context) (os string, arch string, err error)
	CheckDocker(ctx context.Context) (available bool, version string, err error)
	CheckCompose(ctx context.Context) (available bool, version string, err error)
	CheckResources(ctx context.Context) (memBytes uint64, diskBytes uint64, err error)
	CheckPort(ctx context.Context, port int) (available bool, err error)
	CheckRegistry(ctx context.Context) (reachable bool, err error)
}
