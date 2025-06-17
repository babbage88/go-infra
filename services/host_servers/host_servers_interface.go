package host_servers

import (
	"context"
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// HostServer represents a server that can host containers, VMs, or databases
type HostServer struct {
	ID                   uuid.UUID  `json:"id"`
	Hostname             string     `json:"hostname"`
	IPAddress            netip.Addr `json:"ip_address"`
	Username             *string    `json:"username,omitempty"`
	SSHKeyID             *uuid.UUID `json:"ssh_key_id,omitempty"`
	SudoPasswordSecretID *uuid.UUID `json:"sudo_password_secret_id,omitempty"`
	IsContainerHost      bool       `json:"is_container_host"`
	IsVmHost             bool       `json:"is_vm_host"`
	IsVirtualMachine     bool       `json:"is_virtual_machine"`
	IsDbHost             bool       `json:"is_db_host"`
	CreatedAt            time.Time  `json:"created_at"`
	LastModified         time.Time  `json:"last_modified"`
}

// HostServerProvider defines the interface for host server operations
type HostServerProvider interface {
	// CreateHostServer creates a new host server
	CreateHostServer(ctx context.Context, req CreateHostServerRequest) (*HostServer, error)

	// GetHostServer retrieves a host server by ID
	GetHostServer(ctx context.Context, id uuid.UUID) (*HostServer, error)

	// GetHostServerByHostname retrieves a host server by hostname
	GetHostServerByHostname(ctx context.Context, hostname string) (*HostServer, error)

	// GetHostServerByIP retrieves a host server by IP address
	GetHostServerByIP(ctx context.Context, ip netip.Addr) (*HostServer, error)

	// GetAllHostServers retrieves all host servers
	GetAllHostServers(ctx context.Context) ([]HostServer, error)

	// UpdateHostServer updates an existing host server
	UpdateHostServer(ctx context.Context, id uuid.UUID, req UpdateHostServerRequest) (*HostServer, error)

	// DeleteHostServer deletes a host server
	DeleteHostServer(ctx context.Context, id uuid.UUID) error
}
