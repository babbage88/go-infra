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

type SshKeySecretProvider interface {
	StoreSshKeySecret(plaintextSecret string, userId, appId uuid.UUID, expiry time.Time) error
	CreateSshKey(sshKey *NewSshKeyRequest) NewSshKeyResult
}
