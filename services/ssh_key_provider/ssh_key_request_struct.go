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
