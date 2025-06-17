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
	// required: true
	// example: admin
	Username string `json:"username" validate:"required"`

	// SSH key ID for authentication
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174000
	SSHKeyID uuid.UUID `json:"ssh_key_id" validate:"required"`

	// Optional sudo password secret ID
	// required: false
	// example: 123e4567-e89b-12d3-a456-426614174001
	SudoPasswordSecretID *uuid.UUID `json:"sudo_password_secret_id,omitempty"`

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
}

// swagger:parameters UpdateHostServer
// @Description Request to update an existing host server
type UpdateHostServerRequestWrapper struct {
	// in: path
	ID uuid.UUID `json:"id"`
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

	// Optional sudo password secret ID
	// required: false
	// example: 123e4567-e89b-12d3-a456-426614174001
	SudoPasswordSecretID *uuid.UUID `json:"sudo_password_secret_id,omitempty"`

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
}

// swagger:model HostServerResponse
// @Description Response containing a single host server
type HostServerResponse struct {
	ID                   uuid.UUID  `json:"id"`
	Hostname             string     `json:"hostname"`
	IPAddress            netip.Addr `json:"ip_address"`
	Username             string     `json:"username"`
	SSHKeyID             uuid.UUID  `json:"ssh_key_id"`
	SudoPasswordSecretID *uuid.UUID `json:"sudo_password_secret_id,omitempty"`
	IsContainerHost      bool       `json:"is_container_host"`
	IsVmHost             bool       `json:"is_vm_host"`
	IsVirtualMachine     bool       `json:"is_virtual_machine"`
	IsDbHost             bool       `json:"is_db_host"`
	CreatedAt            time.Time  `json:"created_at"`
	LastModified         time.Time  `json:"last_modified"`
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
	ID uuid.UUID `json:"id"`
}
