package ssh_key_provider

import (
	"time"

	"github.com/google/uuid"
)

type NewSshKeyRequest struct {
	UserID       uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	PublicKey    string    `json:"publicKey"`
	PrivateKey   string    `json:"privateKey"`
	HostServerId uuid.UUID `json:"hostServerId"`
	KeyType      string    `json:"keyType"`
	CreatedAt    time.Time `json:"createdAt"`
	LastModified time.Time `json:"lastModified"`
}

type NewSshKeyResult struct {
	SshKeyId        uuid.UUID `json:"id"`
	PrivKeySecretId uuid.UUID `json:"privKeySecretId"`
	UserId          uuid.UUID `json:"userId"`
	Error           error     `json:"error"`
}

// swagger:model SshKeyListItem
type SshKeyListItem struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	PublicKey    string    `json:"publicKey"`
	KeyType      string    `json:"keyType"`
	OwnerUserID  uuid.UUID `json:"ownerUserId"`
	CreatedAt    time.Time `json:"createdAt"`
	LastModified time.Time `json:"lastModified"`
}

type CreateSshKeyHostMappingResult struct {
	ID                 uuid.UUID `json:"id"`
	SshKeyID           uuid.UUID `json:"sshKeyId"`
	HostServerID       uuid.UUID `json:"hostServerId"`
	UserID             uuid.UUID `json:"userId"`
	HostserverUsername string    `json:"hostserverUsername"`
	CreatedAt          time.Time `json:"createdAt"`
	LastModified       time.Time `json:"lastModified"`
	Error              error     `json:"error"`
}

type UpdateSshKeyHostMappingResult struct {
	ID                 uuid.UUID `json:"id"`
	SshKeyID           uuid.UUID `json:"sshKeyId"`
	HostServerID       uuid.UUID `json:"hostServerId"`
	UserID             uuid.UUID `json:"userId"`
	HostserverUsername string    `json:"hostserverUsername"`
	CreatedAt          time.Time `json:"createdAt"`
	LastModified       time.Time `json:"lastModified"`
	Error              error     `json:"error"`
}

type SshKeySecretProvider interface {
	CreateSshKey(sshKey *NewSshKeyRequest) NewSshKeyResult
	DeleteSShKeyAndSecret(sshKeyId uuid.UUID) error
	GetSshKeysByUserId(userId uuid.UUID) ([]SshKeyListItem, error)

	// SSH Key Host Mapping CRUD operations
	CreateSshKeyHostMapping(mapping *CreateSshKeyHostMappingRequest) CreateSshKeyHostMappingResult
	GetSshKeyHostMappingById(id uuid.UUID) (*CreateSshKeyHostMappingResult, error)
	GetSshKeyHostMappingsByUserId(userId uuid.UUID) ([]CreateSshKeyHostMappingResult, error)
	GetSshKeyHostMappingsByHostId(hostId uuid.UUID) ([]CreateSshKeyHostMappingResult, error)
	GetSshKeyHostMappingsByKeyId(keyId uuid.UUID) ([]CreateSshKeyHostMappingResult, error)
	UpdateSshKeyHostMapping(mapping *UpdateSshKeyHostMappingRequest) UpdateSshKeyHostMappingResult
	DeleteSshKeyHostMapping(id uuid.UUID) error
	DeleteSshKeyHostMappingsBySshKeyId(sshKeyId uuid.UUID) error
}
