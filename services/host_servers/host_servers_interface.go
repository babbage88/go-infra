package host_servers

import (
	"context"
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// HostServer represents a server that can host containers, VMs, or databases
type HostServer struct {
	ID                   uuid.UUID        `json:"id"`
	Hostname             string           `json:"hostname"`
	IPAddress            netip.Addr       `json:"ip_address"`
	Username             *string          `json:"username,omitempty"`
	SSHKeyID             *uuid.UUID       `json:"ssh_key_id,omitempty"`
	SudoPasswordSecretID *uuid.UUID       `json:"sudo_password_secret_id,omitempty"`
	IsContainerHost      bool             `json:"is_container_host"`
	IsVmHost             bool             `json:"is_vm_host"`
	IsVirtualMachine     bool             `json:"is_virtual_machine"`
	IsDbHost             bool             `json:"is_db_host"`
	HostServerTypes      []HostServerType `json:"host_server_types,omitempty"`
	PlatformTypes        []PlatformType   `json:"platform_types,omitempty"`
	CreatedAt            time.Time        `json:"created_at"`
	LastModified         time.Time        `json:"last_modified"`
}

// swagger:model HostServerType
// @Description A type/category of host server
type HostServerType struct {
	// Unique identifier for the host server type
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174000
	ID uuid.UUID `json:"id"`

	// Name of the host server type
	// required: true
	// example: Database Server
	Name string `json:"name"`

	// Last modification timestamp
	// required: true
	// example: 2024-01-15T10:30:00Z
	LastModified time.Time `json:"last_modified"`
}

// swagger:model PlatformType
// @Description A specific platform or service running on a host server
type PlatformType struct {
	// Unique identifier for the platform type
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174001
	ID uuid.UUID `json:"id"`

	// Name of the platform type
	// required: true
	// example: Docker Host
	Name string `json:"name"`

	// Last modification timestamp
	// required: true
	// example: 2024-01-15T10:30:00Z
	LastModified time.Time `json:"last_modified"`
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

	// GetAllHostServerTypes retrieves all available host server types
	GetAllHostServerTypes(ctx context.Context) ([]HostServerType, error)

	// GetAllPlatformTypes retrieves all available platform types
	GetAllPlatformTypes(ctx context.Context) ([]PlatformType, error)

	// CreateHostServerTypeMapping creates a mapping between a host server and a host server type
	CreateHostServerTypeMapping(ctx context.Context, hostServerID, hostServerTypeID uuid.UUID) error

	// CreatePlatformTypeMapping creates a mapping between a host server, platform type, and host server type
	CreatePlatformTypeMapping(ctx context.Context, hostServerID, platformTypeID, hostServerTypeID uuid.UUID) error
}
