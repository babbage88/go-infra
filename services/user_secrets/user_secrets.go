package user_secrets

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserSecretProvider interface {
	StoreSecret(plaintextSecret string, userId, appId uuid.UUID) error
	RetrieveSecret(userId, appId uuid.UUID)
}

type PgUserSecretStore struct {
	DbConn *pgxpool.Pool `json:"dbConn"`
}

type PgEncrytpedAuthToken struct {
	UserId        uuid.UUID                      `json:"userId"`
	ApplicationId uuid.UUID                      `json:"applicationId"`
	UserSecret    *EncryptedUserSecretsAES256GCM `json:"userSecret"`
}
