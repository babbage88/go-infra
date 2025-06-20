package ssh_key_provider

import (
	"github.com/google/uuid"
)

// swagger:parameters createSshKey
type CreateSshKeyRequestWrapper struct {
	// in:body
	Body CreateSshKeyRequest `json:"body"`
}

// swagger:model CreateSshKeyRequest
type CreateSshKeyRequest struct {
	// Name of the SSH key
	// required: true
	Name string `json:"name"`

	// Description of the SSH key
	// required: false
	Description string `json:"description"`

	// Public key in OpenSSH format
	// required: true
	PublicKey string `json:"publicKey"`

	// Private key in PEM format
	// required: true
	PrivateKey string `json:"privateKey"`

	// Type of the SSH key (e.g., rsa, ed25519)
	// required: true
	KeyType string `json:"keyType"`

	// Optional host server ID to associate the key with
	// required: false
	HostServerId *uuid.UUID `json:"hostServerId,omitempty"`
}

// swagger:response CreateSshKeyResponse
type CreateSshKeyResponseWrapper struct {
	// in:body
	Body CreateSshKeyResponse `json:"body"`
}

// swagger:model CreateSshKeyResponse
type CreateSshKeyResponse struct {
	// ID of the created SSH key
	// required: true
	SshKeyId uuid.UUID `json:"sshKeyId"`

	// ID of the stored private key secret
	// required: true
	PrivKeySecretId uuid.UUID `json:"privKeySecretId"`

	// ID of the user who owns the key
	// required: true
	UserId uuid.UUID `json:"userId"`

	// Error message if the operation failed
	// required: false
	Error string `json:"error,omitempty"`
}

// SSH Key Host Mapping CRUD Request/Response structs

// swagger:parameters createSshKeyHostMapping
type CreateSshKeyHostMappingRequestWrapper struct {
	// in:body
	Body CreateSshKeyHostMappingRequest `json:"body"`
}

// swagger:response CreateSshKeyHostMappingResponse
type CreateSshKeyHostMappingResponseWrapper struct {
	// in:body
	Body CreateSshKeyHostMappingResponse `json:"body"`
}

// swagger:model CreateSshKeyHostMappingResponse
type CreateSshKeyHostMappingResponse struct {
	// ID of the created mapping
	// required: true
	ID uuid.UUID `json:"id"`

	// ID of the SSH key
	// required: true
	SshKeyID uuid.UUID `json:"sshKeyId"`

	// ID of the host server
	// required: true
	HostServerID uuid.UUID `json:"hostServerId"`

	// ID of the user
	// required: true
	UserID uuid.UUID `json:"userId"`

	// Username on the host server
	// required: true
	HostserverUsername string `json:"hostserverUsername"`

	// Creation timestamp
	// required: true
	CreatedAt string `json:"createdAt"`

	// Last modification timestamp
	// required: true
	LastModified string `json:"lastModified"`

	// Error message if the operation failed
	// required: false
	Error string `json:"error,omitempty"`
}

// swagger:parameters getSshKeyHostMappingById
type GetSshKeyHostMappingByIdRequestWrapper struct {
	// ID of the SSH key host mapping to retrieve
	// in: path
	// required: true
	ID string `json:"id"`
}

// swagger:response GetSshKeyHostMappingByIdResponse
type GetSshKeyHostMappingByIdResponseWrapper struct {
	// in:body
	Body CreateSshKeyHostMappingResponse `json:"body"`
}

// swagger:parameters getSshKeyHostMappingsByUserId
type GetSshKeyHostMappingsByUserIdRequestWrapper struct {
	// ID of the user to get mappings for
	// in: path
	// required: true
	UserID string `json:"userId"`
}

// swagger:response GetSshKeyHostMappingsByUserIdResponse
type GetSshKeyHostMappingsByUserIdResponseWrapper struct {
	// in:body
	Body []CreateSshKeyHostMappingResponse `json:"body"`
}

// swagger:parameters getSshKeyHostMappingsByHostId
type GetSshKeyHostMappingsByHostIdRequestWrapper struct {
	// ID of the host server to get mappings for
	// in: path
	// required: true
	HostID string `json:"hostId"`
}

// swagger:response GetSshKeyHostMappingsByHostIdResponse
type GetSshKeyHostMappingsByHostIdResponseWrapper struct {
	// in:body
	Body []CreateSshKeyHostMappingResponse `json:"body"`
}

// swagger:parameters getSshKeyHostMappingsByKeyId
type GetSshKeyHostMappingsByKeyIdRequestWrapper struct {
	// ID of the SSH key to get mappings for
	// in: path
	// required: true
	KeyID string `json:"keyId"`
}

// swagger:response GetSshKeyHostMappingsByKeyIdResponse
type GetSshKeyHostMappingsByKeyIdResponseWrapper struct {
	// in:body
	Body []CreateSshKeyHostMappingResponse `json:"body"`
}

// swagger:parameters updateSshKeyHostMapping
type UpdateSshKeyHostMappingRequestWrapper struct {
	// ID of the SSH key host mapping to update
	// in: path
	// required: true
	ID string `json:"id"`

	// in:body
	Body UpdateSshKeyHostMappingRequest `json:"body"`
}

// swagger:model UpdateSshKeyHostMappingRequest
type UpdateSshKeyHostMappingRequest struct {
	// ID of the SSH key host mapping to update
	// required: true
	ID uuid.UUID `json:"id"`

	// Username to use on the host server
	// required: true
	HostserverUsername string `json:"hostserverUsername"`
}

// swagger:response UpdateSshKeyHostMappingResponse
type UpdateSshKeyHostMappingResponseWrapper struct {
	// in:body
	Body CreateSshKeyHostMappingResponse `json:"body"`
}

// swagger:parameters deleteSshKeyHostMapping
type DeleteSshKeyHostMappingRequestWrapper struct {
	// ID of the SSH key host mapping to delete
	// in: path
	// required: true
	ID string `json:"id"`
}

// swagger:response DeleteSshKeyHostMappingResponse
type DeleteSshKeyHostMappingResponseWrapper struct {
	// in:body
	Body DeleteSshKeyHostMappingResponse `json:"body"`
}

// swagger:model DeleteSshKeyHostMappingResponse
type DeleteSshKeyHostMappingResponse struct {
	// Success message
	// required: true
	Message string `json:"message"`
}

// SSH Key Host Mapping CRUD operations

// swagger:model CreateSshKeyHostMappingRequest
type CreateSshKeyHostMappingRequest struct {
	// ID of the SSH key to map
	// required: true
	SshKeyID uuid.UUID `json:"sshKeyId"`

	// ID of the host server to map to
	// required: true
	HostServerID uuid.UUID `json:"hostServerId"`

	// ID of the user who owns the mapping
	// required: true
	UserID uuid.UUID `json:"userId"`

	// Username to use on the host server
	// required: true
	HostserverUsername string `json:"hostserverUsername"`
}
