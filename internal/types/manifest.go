package types

import "time"

// 1. ADVANCED: Use custom types and constants for Enums
// This prevents typos like "pending" vs "Pending" in your state machine.
type MigrationStatus string

const (
	StatusPending     MigrationStatus = "Pending"
	StatusCompressing MigrationStatus = "Compressing"
	StatusSending     MigrationStatus = "Sending"
	StatusExtracting  MigrationStatus = "Extracting"
	StatusSuccess     MigrationStatus = "Success"
	StatusFailed      MigrationStatus = "Failed"
)

// MigrationEvent represents a progress update for a batch.
// Useful for WebSockets, SSE (Server-Sent Events), or UI polling.
type MigrationEvent struct {
	ContainerID string          `json:"container_id"`
	Status      MigrationStatus `json:"status"`
	Progress    int             `json:"progress"`        // 0-100
	Error       string          `json:"error,omitempty"` // Omit if no error
	Timestamp   time.Time       `json:"timestamp"`       // ADVANCED: Crucial for client-side sorting/UI
}

// ContainerManifest is the DNA sent to the receiver.
type ContainerManifest struct {
	Name  string   `json:"name"`
	Image string   `json:"image"`
	Env   []string `json:"env,omitempty"`

	// ADVANCED: Missing crucial Docker configurations
	Cmd        []string          `json:"cmd,omitempty"`        // Custom run commands
	Entrypoint []string          `json:"entrypoint,omitempty"` // Custom entrypoints
	Ports      []PortMapping     `json:"ports,omitempty"`      // Port bindings (Crucial for web apps)
	Labels     map[string]string `json:"labels,omitempty"`     // Metadata (e.g. "migrated-by=dockporter")

	Mounts []MountDefinition `json:"mounts,omitempty"`
}

// PortMapping defines how container ports map to the host
type PortMapping struct {
	HostPort      string `json:"host_port"`
	ContainerPort string `json:"container_port"`
	Protocol      string `json:"protocol"` // "tcp" or "udp"
}

// MountDefinition details volume and bind mounts
type MountDefinition struct {
	Type        string `json:"type"`        // ADVANCED: "bind" or "volume"
	Source      string `json:"source"`      // Host path or Volume name
	Destination string `json:"destination"` // Container path
	ReadOnly    bool   `json:"read_only"`   // Prevent accidental writes to read-only mounts
}
