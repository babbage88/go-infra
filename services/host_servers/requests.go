package host_servers

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// swagger:parameters CreateHostServer
// @Description Request to create a new host server
type CreateHostServerRequestWrapper struct {
	// in: body
	Body CreateHostServerRequest `json:"body"`
}

// swagger:model CreateHostServerRequest
type CreateHostServerRequest struct {
	// Hostname of the server
	// required: true
	// example: server-01.example.com
	Hostname string `json:"hostname" validate:"required"`

	// IP address of the server
	// required: true
	// example: 192.168.1.100
	IPAddress netip.Addr `json:"ip_address" validate:"required"`

	// Username for SSH connection
	// required: false
	// example: admin
	Username *string `json:"username,omitempty"`

	// SSH key ID for authentication
	// required: false
	// example: 123e4567-e89b-12d3-a456-426614174000
	SSHKeyID *uuid.UUID `json:"ssh_key_id,omitempty"`

	// Optional sudo password token ID
	// required: false
	// example: 123e4567-e89b-12d3-a456-426614174001
	SudoPasswordTokenID *uuid.UUID `json:"sudo_password_token_id,omitempty"`

	// Whether this server can host containers
	// required: false
	// example: true
	IsContainerHost bool `json:"is_container_host"`

	// Whether this server can host VMs
	// required: false
	// example: false
	IsVmHost bool `json:"is_vm_host"`

	// Whether this server is a virtual machine
	// required: false
	// example: false
	IsVirtualMachine bool `json:"is_virtual_machine"`

	// Whether this server can host databases
	// required: false
	// example: false
	IsDbHost bool `json:"is_db_host"`

	// Host server type IDs that this server supports
	// required: false
	// example: ["123e4567-e89b-12d3-a456-426614174000", "123e4567-e89b-12d3-a456-426614174001"]
	HostServerTypeIDs []uuid.UUID `json:"host_server_type_ids,omitempty"`

	// Platform type IDs that this server supports
	// required: false
	// example: ["123e4567-e89b-12d3-a456-426614174002", "123e4567-e89b-12d3-a456-426614174003"]
	PlatformTypeIDs []uuid.UUID `json:"platform_type_ids,omitempty"`
}

// swagger:parameters UpdateHostServer
// @Description Request to update an existing host server
type UpdateHostServerRequestWrapper struct {
	// in: path
	ID uuid.UUID `json:"ID"`
	// in: body
	Body UpdateHostServerRequest `json:"body"`
}

// swagger:model UpdateHostServerRequest
type UpdateHostServerRequest struct {
	// Hostname of the server
	// required: false
	// example: server-01.example.com
	Hostname *string `json:"hostname,omitempty"`

	// IP address of the server
	// required: false
	// example: 192.168.1.100
	IPAddress *netip.Addr `json:"ip_address,omitempty"`

	// Username for SSH connection
	// required: false
	// example: admin
	Username *string `json:"username,omitempty"`

	// SSH key ID for authentication
	// required: false
	// example: 123e4567-e89b-12d3-a456-426614174000
	SSHKeyID *uuid.UUID `json:"ssh_key_id,omitempty"`

	// Optional sudo password token ID
	// required: false
	// example: 123e4567-e89b-12d3-a456-426614174001
	SudoPasswordTokenID *uuid.UUID `json:"sudo_password_token_id,omitempty"`

	// Whether this server can host containers
	// required: false
	// example: true
	IsContainerHost *bool `json:"is_container_host,omitempty"`

	// Whether this server can host VMs
	// required: false
	// example: false
	IsVmHost *bool `json:"is_vm_host,omitempty"`

	// Whether this server is a virtual machine
	// required: false
	// example: false
	IsVirtualMachine *bool `json:"is_virtual_machine,omitempty"`

	// Whether this server can host databases
	// required: false
	// example: false
	IsDbHost *bool `json:"is_db_host,omitempty"`

	// Host server type IDs that this server supports
	// required: false
	// example: ["123e4567-e89b-12d3-a456-426614174000", "123e4567-e89b-12d3-a456-426614174001"]
	HostServerTypeIDs []uuid.UUID `json:"host_server_type_ids,omitempty"`

	// Platform type IDs that this server supports
	// required: false
	// example: ["123e4567-e89b-12d3-a456-426614174002", "123e4567-e89b-12d3-a456-426614174003"]
	PlatformTypeIDs []uuid.UUID `json:"platform_type_ids,omitempty"`
}

// HostServerResponse represents a host server response.
// swagger:model
type HostServerResponse struct {
	ID                  uuid.UUID        `json:"id"`
	Hostname            string           `json:"hostname"`
	IPAddress           netip.Addr       `json:"ip_address"`
	Username            *string          `json:"username,omitempty"`
	SSHKeyID            *uuid.UUID       `json:"ssh_key_id,omitempty"`
	SudoPasswordTokenID *uuid.UUID       `json:"sudo_password_token_id,omitempty"`
	IsContainerHost     bool             `json:"is_container_host"`
	IsVmHost            bool             `json:"is_vm_host"`
	IsVirtualMachine    bool             `json:"is_virtual_machine"`
	IsDbHost            bool             `json:"is_db_host"`
	HostServerTypes     []HostServerType `json:"host_server_types,omitempty"`
	PlatformTypes       []PlatformType   `json:"platform_types,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	LastModified        time.Time        `json:"last_modified"`
}

// swagger:model HostServersResponse
// @Description Response containing multiple host servers
type HostServersResponse []HostServerResponse

// swagger:response HostServersResponse
type HostServersResponseWrapper struct {
	// in: body
	Body []HostServerResponse `json:"body"`
}

// swagger:response HostServerResponse
type HostServerResponseWrapper struct {
	// in: body
	Body HostServerResponse `json:"body"`
}

// swagger:parameters DeleteHostServer
// @Description Request to delete a host server
type DeleteHostServerRequestWrapper struct {
	// in: path
	ID uuid.UUID `json:"ID"`
}

// swagger:parameters GetAllHostServerTypes
// @Description Request to get all host server types
type GetAllHostServerTypesRequestWrapper struct {
	// No parameters needed for this endpoint
}

// swagger:response GetAllHostServerTypesResponse
type GetAllHostServerTypesResponseWrapper struct {
	// in: body
	Body []HostServerType `json:"body"`
}

// swagger:parameters GetAllPlatformTypes
// @Description Request to get all platform types
type GetAllPlatformTypesRequestWrapper struct {
	// No parameters needed for this endpoint
}

// swagger:response GetAllPlatformTypesResponse
type GetAllPlatformTypesResponseWrapper struct {
	// in: body
	Body []PlatformType `json:"body"`
}

// swagger:parameters CreateHostServerTypeMapping
// @Description Request to create a host server type mapping
type CreateHostServerTypeMappingRequestWrapper struct {
	// in:body
	Body CreateHostServerTypeMappingRequest `json:"body"`
}

// swagger:model CreateHostServerTypeMappingRequest
type CreateHostServerTypeMappingRequest struct {
	// Host server ID
	// required: true
	HostServerId uuid.UUID `json:"hostServerId"`
	// Host server type ID
	// required: true
	HostServerTypeId uuid.UUID `json:"hostServerTypeId"`
}

// swagger:response CreateHostServerTypeMappingResponse
type CreateHostServerTypeMappingResponseWrapper struct {
	// in:body
	Body struct {
		Success bool `json:"success"`
	} `json:"body"`
}

// swagger:parameters CreatePlatformTypeMapping
// @Description Request to create a platform type mapping
type CreatePlatformTypeMappingRequestWrapper struct {
	// in:body
	Body CreatePlatformTypeMappingRequest `json:"body"`
}

// swagger:model CreatePlatformTypeMappingRequest
type CreatePlatformTypeMappingRequest struct {
	// Host server ID
	// required: true
	HostServerId uuid.UUID `json:"hostServerId"`
	// Platform type ID
	// required: true
	PlatformTypeId uuid.UUID `json:"platformTypeId"`
	// Host server type ID
	// required: true
	HostServerTypeId uuid.UUID `json:"hostServerTypeId"`
}

// swagger:response CreatePlatformTypeMappingResponse
type CreatePlatformTypeMappingResponseWrapper struct {
	// in:body
	Body struct {
		Success bool `json:"success"`
	} `json:"body"`
}
