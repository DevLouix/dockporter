package dockerops

import (
	"context" // Add this
	"fmt"
	"github.com/docker/docker/client"
)

type DockerManager struct {
	cli *client.Client
}

// Update the signature to include (ctx context.Context)
func NewDockerManager(ctx context.Context) (*DockerManager, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv, 
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %w", err)
	}

	// This is why we need the context: to verify the connection works
	_, err = cli.Ping(ctx)
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("docker daemon unreachable: %w", err)
	}

	return &DockerManager{cli: cli}, nil
}

func (dm *DockerManager) Close() error {
	return dm.cli.Close()
}