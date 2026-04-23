package orchestrator

import (
	"context"
	"fmt"
	"log"
	"strings"

	// Alias your internal types
	shift "dockporter/internal/types"

	"dockporter/internal/api"
	"dockporter/internal/dockerops"
)

// ShipContainer now accepts remoteToken to satisfy security requirements
func ShipContainer(ctx context.Context, dm *dockerops.DockerManager, containerID string, remoteAddr string, remoteToken string) error {

	// 1. Inspect the container (Concrete type: dockertypes.ContainerJSON)
	info, err := dm.InspectContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to inspect container: %w", err)
	}

	// 2. Extract DNA into Manifest
	manifest := shift.ContainerManifest{
		Name:       strings.TrimPrefix(info.Name, "/"),
		Image:      info.Config.Image,
		Env:        info.Config.Env,
		Cmd:        info.Config.Cmd,
		Entrypoint: info.Config.Entrypoint,
		Labels:     info.Config.Labels,
	}

	// Extract Ports
	for port, bindings := range info.HostConfig.PortBindings {
		for _, b := range bindings {
			manifest.Ports = append(manifest.Ports, shift.PortMapping{
				ContainerPort: port.Port(),
				Protocol:      port.Proto(),
				HostPort:      b.HostPort,
			})
		}
	}

	// 3. Locate Volume Data
	if len(info.Mounts) == 0 {
		return fmt.Errorf("container %s has no mounts to migrate", containerID)
	}
	sourcePath := info.Mounts[0].Source
	log.Printf("📦 Source found: %s. Starting stream to %s...", sourcePath, remoteAddr)

	// 4. Start Export Pipeline
	stream, errChan := dm.ExportVolumeStream(sourcePath)

	// 5. Send over the wire (Now passing the token!)
	netErr := api.SendMigrationStream(ctx, remoteAddr, remoteToken, manifest, stream)

	// 6. Sync Errors
	walkErr := <-errChan

	if netErr != nil {
		return fmt.Errorf("migration upload failed: %w", netErr)
	}
	if walkErr != nil {
		return fmt.Errorf("local compression failed: %w", walkErr)
	}

	return nil
}
