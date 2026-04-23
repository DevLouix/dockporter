package dockerops

import (
	"context"
	"fmt"

	// Notice the space after the alias
	manifest "dockporter/internal/types"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
)

// ListContainers asks Docker for ALL containers (running, stopped, paused).
func (dm *DockerManager) ListContainers(ctx context.Context) ([]types.Container, error) {
	// ADVANCED: All: true is critical for a control panel!
	return dm.cli.ContainerList(ctx, container.ListOptions{All: true})
}

func (dm *DockerManager) InspectContainer(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return dm.cli.ContainerInspect(ctx, containerID)
}

func (dm *DockerManager) StopContainer(ctx context.Context, containerID string) error {
	return dm.cli.ContainerStop(ctx, containerID, container.StopOptions{})
}

func (dm *DockerManager) StartContainer(ctx context.Context, containerID string) error {
	return dm.cli.ContainerStart(ctx, containerID, container.StartOptions{})
}

// ADVANCED: Added Remove Container (Supports forceful deletion of running containers)
func (dm *DockerManager) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return dm.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force:         force,
		RemoveVolumes: false, // Set to true if you want to wipe data on delete
	})
}

// ADVANCED: Added Rename Container
func (dm *DockerManager) RenameContainer(ctx context.Context, containerID string, newName string) error {
	return dm.cli.ContainerRename(ctx, containerID, newName)
}

// CreateMigratedContainer translates the ContainerManifest into Docker SDK configurations.
func (dm *DockerManager) CreateMigratedContainer(
	ctx context.Context,
	manifest manifest.ContainerManifest,
	restoredVolumePath string,
) (container.CreateResponse, error) {

	// 1. Map exposed ports to host ports
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	for _, p := range manifest.Ports {
		// e.g., "80/tcp"
		portStr := nat.Port(fmt.Sprintf("%s/%s", p.ContainerPort, p.Protocol))

		exposedPorts[portStr] = struct{}{}
		portBindings[portStr] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0", // Bind to all interfaces
				HostPort: p.HostPort,
			},
		}
	}

	// 2. Prepare the Container's Internal Configuration
	config := &container.Config{
		Image:        manifest.Image,
		Env:          manifest.Env,
		Cmd:          manifest.Cmd,
		Entrypoint:   manifest.Entrypoint,
		ExposedPorts: exposedPorts,
		Labels:       manifest.Labels,
	}

	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}
	config.Labels["managed-by"] = "dockporter"

	// 3. Prepare Host Configuration (Hardware, Networking, Mounts)
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: restoredVolumePath, // The unpacked volume data
				Target: "/data",            // Adjust based on where your container expects data
			},
		},
	}

	// Append any additional mounts declared in the manifest
	for _, m := range manifest.Mounts {
		mntType := mount.TypeVolume
		if m.Type == "bind" {
			mntType = mount.TypeBind
		}

		hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
			Type:     mntType,
			Source:   m.Source,
			Target:   m.Destination,
			ReadOnly: m.ReadOnly,
		})
	}

	// 4. Send creation request to Docker Daemon
	return dm.cli.ContainerCreate(
		ctx,
		config,
		hostConfig,
		nil, // Network config (nil = default bridge)
		nil, // Platform config
		manifest.Name,
	)
}
