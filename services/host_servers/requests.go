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
	// Host server ID
	// in: path
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174000
	ID string `json:"ID"`
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

	// Clear IP address from the server
	// required: false
	// example: true
	ClearIPAddress *bool `json:"clear_ip_address,omitempty"`

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
	IPAddress           *netip.Addr      `json:"ip_address,omitempty"`
	Username            *string          `json:"username,omitempty"`
	SSHKeyID            *uuid.UUID       `json:"ssh_key_id,omitempty"`
	SudoPasswordTokenID *uuid.UUID       `json:"sudo_password_token_id,omitempty"`
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
	// Host server ID
	// in: path
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174000
	ID string `json:"ID"`
}

// swagger:parameters GetHostServer
// @Description Request to get a host server by ID
type GetHostServerRequestWrapper struct {
	// Host server ID
	// in: path
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174000
	ID string `json:"ID"`
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

// swagger:parameters CreatePlatformType
// @Description Request to create a new platform type by name
type CreatePlatformTypeRequest struct {
	// in: path
	NAME string `json:"NAME"`
}

// swagger:parameters CreatePlatformType
// @Description Request to create a new host server type by name
type CreateHostServerTypeRequest struct {
	// in: path
	NAME string `json:"NAME"`
}

// swagger:response CreatePlatformTypeResponse
type CreatePlatformTypeResponse struct {
	// in:body
	Body struct {
		Id   uuid.UUID `json:"platformId"`
		Name string    `json:"name"`
	} `json:"body"`
}

// swagger:response CreateHostServerTypeResponse
type CreateHostServerTypeResponse struct {
	// in:body
	Body struct {
		Id   uuid.UUID `json:"hostServerId"`
		Name string    `json:"name"`
	} `json:"body"`
}

// swagger:parameters CreateHostServerType
// @Description Request to create a new host server type
type CreateHostServerTypeBodyRequest struct {
	// in: body
	Body CreateHostServerTypeBodyRequestBody `json:"body"`
}

// swagger:model CreateHostServerTypeBodyRequest
type CreateHostServerTypeBodyRequestBody struct {
	// Name of the host server type
	// required: true
	// example: Database Server
	Name string `json:"name" validate:"required"`
}

// swagger:parameters CreatePlatformTypeBody
// @Description Request to create a new platform type
type CreatePlatformTypeBodyRequest struct {
	// in: body
	Body CreatePlatformTypeBodyRequestBody `json:"body"`
}

// swagger:model CreatePlatformTypeBodyRequest
type CreatePlatformTypeBodyRequestBody struct {
	// Name of the platform type
	// required: true
	// example: Docker Host
	Name string `json:"name" validate:"required"`
}

// swagger:response HostServerTypeResponse
type HostServerTypeResponse struct {
	// in: body
	Body HostServerType `json:"body"`
}

// swagger:response PlatformTypeResponse
type PlatformTypeResponse struct {
	// in: body
	Body PlatformType `json:"body"`
}

// swagger:parameters GetHostServerTypeById
// @Description Request to get a host server type by ID
type GetHostServerTypeByIdRequest struct {
	// Host server type ID
	// in: path
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174000
	ID string `json:"ID"`
}

// swagger:parameters GetHostServerTypeByName
// @Description Request to get a host server type by name
type GetHostServerTypeByNameRequest struct {
	// in: path
	Name string `json:"name"`
}

// swagger:parameters UpdateHostServerType
// @Description Request to update a host server type
type UpdateHostServerTypeRequest struct {
	// Host server type ID
	// in: path
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174000
	ID string `json:"ID"`
	// in: body
	Body UpdateHostServerTypeBody `json:"body"`
}

// swagger:model UpdateHostServerTypeBody
type UpdateHostServerTypeBody struct {
	// Name of the host server type
	// required: false
	// example: Database Server
	Name *string `json:"name,omitempty"`
}

// swagger:parameters DeleteHostServerType
// @Description Request to delete a host server type
type DeleteHostServerTypeRequest struct {
	// Host server type ID
	// in: path
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174000
	ID string `json:"ID"`
}

// swagger:parameters GetPlatformTypeById
// @Description Request to get a platform type by ID
type GetPlatformTypeByIdRequest struct {
	// Platform type ID
	// in: path
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174001
	ID string `json:"ID"`
}

// swagger:parameters GetPlatformTypeByName
// @Description Request to get a platform type by name
type GetPlatformTypeByNameRequest struct {
	// in: path
	Name string `json:"name"`
}

// swagger:parameters UpdatePlatformType
// @Description Request to update a platform type
type UpdatePlatformTypeRequest struct {
	// Platform type ID
	// in: path
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174001
	ID string `json:"ID"`
	// in: body
	Body UpdatePlatformTypeBody `json:"body"`
}

// swagger:model UpdatePlatformTypeBody
type UpdatePlatformTypeBody struct {
	// Name of the platform type
	// required: false
	// example: Docker Host
	Name *string `json:"name,omitempty"`
}

// swagger:parameters DeletePlatformType
// @Description Request to delete a platform type
type DeletePlatformTypeRequest struct {
	// Platform type ID
	// in: path
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174001
	ID string `json:"ID"`
}
