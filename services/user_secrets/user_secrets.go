package user_secrets

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"time"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RetrievedUserSecret struct {
	Reader            io.Reader
	ExternalAuthToken *ExternalApplicationAuthToken
}

// swagger:model UserSecretEntry
type UserSecretEntry struct {
	AppInfo        ExternalApplicationInfo   `json:"appInfo"`
	SecretMetadata ExternalAppSecretMetadata `json:"secretMetadata"`
}

// swagger:model ExternalApplicationInfo
type ExternalApplicationInfo struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	UrlEndpoint string    `json:"url"`
}

// swagger:model ExternalAppSecretMetadata
type ExternalAppSecretMetadata struct {
	Id         uuid.UUID `json:"id"`
	UserId     uuid.UUID `json:"userId"`
	Expiration time.Time `json:"expiry"`
	CreatedAt  time.Time `json:"createdAt"`
}

type UserSecretProvider interface {
	StoreSecret(plaintextSecret string, userId, appId uuid.UUID, expiry time.Time) (uuid.UUID, error)
	RetrieveSecret(secretId uuid.UUID) (*RetrievedUserSecret, error)
	GetUserSecretEntries(userId uuid.UUID) ([]UserSecretEntry, error)
	GetUserSecretEntriesByAppId(userId uuid.UUID, appId uuid.UUID) ([]UserSecretEntry, error)
	GetUserSecretEntriesByAppName(userId uuid.UUID, appName string) ([]UserSecretEntry, error)
	DeleteSecret(secretId uuid.UUID) error
}

// Implementing UserSecretProvider for scenarios where the user supplied secret has only one value that needs to be encrypted
// such as JWT a single bearer token such as a cloudflare token
type PgUserSecretStore struct {
	DbConn *pgxpool.Pool `json:"dbConn"`
}

type PgEncrytpedSecret struct {
	UserId        uuid.UUID                      `json:"userId"`
	ApplicationId uuid.UUID                      `json:"applicationId"`
	UserSecret    *EncryptedUserSecretsAES256GCM `json:"userSecret"`
}

func (p *PgUserSecretStore) StoreSecret(plaintextSecret string, userId, appId uuid.UUID, expiry time.Time) (uuid.UUID, error) {
	userCipherText, err := Encrypt(plaintextSecret)
	if err != nil {
		slog.Error("Error encrypting user secret", slog.String("Error", err.Error()))
		return uuid.Nil, err
	}

	userSecret := PgEncrytpedSecret{
		UserId:        userId,
		ApplicationId: appId,
		UserSecret:    &userCipherText,
	}

	jsonData, err := json.Marshal(userSecret)
	if err != nil {
		slog.Error("Failed to marshal encrypted secret to JSON", slog.String("error", err.Error()))
		return uuid.Nil, err
	}

	if expiry.IsZero() {
		today := time.Now()
		expiry = today.AddDate(0, 0, 31)
	}

	qry := infra_db_pg.New(p.DbConn)
	params := infra_db_pg.InsertExternalAuthTokenParams{
		UserID:        userId,
		ExternalAppID: appId,
		Token:         jsonData,
		Expiration:    pgtype.Timestamptz{Time: expiry, Valid: true},
	}
	secretId, err := qry.InsertExternalAuthToken(context.Background(), params)
	if err != nil {
		return uuid.Nil, err
	}
	return secretId, nil
}

func (p *PgUserSecretStore) RetrieveSecret(secretId uuid.UUID) (*RetrievedUserSecret, error) {
	qry := infra_db_pg.New(p.DbConn)
	record, err := qry.GetExternalAuthTokenById(context.Background(), secretId)
	if err != nil {
		slog.Error("Error retrieving user secret from database", slog.String("error", err.Error()))
		return nil, err
	}

	var stored PgEncrytpedSecret
	err = json.Unmarshal(record.Token, &stored)
	if err != nil {
		slog.Error("Failed to unmarshal encrypted secret", slog.String("error", err.Error()))
		return nil, err
	}

	plaintext, err := stored.UserSecret.Decrypt()
	if err != nil {
		slog.Error("Failed to decrypt secret", slog.String("error", err.Error()))
		return nil, err
	}

	daoExtSecret := ExternalApplicationAuthToken{
		Id:                    record.ID,
		UserID:                record.UserID,
		ExternalApplicationId: record.ExternalAppID,
		Expiration:            record.Expiration.Time,
		Token:                 plaintext,
	}

	return &RetrievedUserSecret{
		Reader:            bytes.NewReader(plaintext),
		ExternalAuthToken: &daoExtSecret,
	}, nil
}

func (p *PgUserSecretStore) DeleteSecret(secretId uuid.UUID) error {
	qry := infra_db_pg.New(p.DbConn)
	result := qry.DeleteExternalAuthTokenById(context.Background(), secretId)
	return result
}

func (p *PgUserSecretStore) GetUserSecretEntries(userId uuid.UUID) ([]UserSecretEntry, error) {
	userSecrets := make([]UserSecretEntry, 0)
	qry := infra_db_pg.New(p.DbConn)
	tokens, err := qry.GetUserSecretsByUserId(context.Background(), userId)
	if err != nil {
		slog.Error("Error retrieving token metadata from database", slog.String("error", err.Error()))
		return userSecrets, err
	}
	for _, entryDb := range tokens {
		var entry UserSecretEntry
		entry.ParseExternalAppSecretMetadataFromDb(entryDb)
		userSecrets = append(userSecrets, entry)
	}

	return userSecrets, nil
}

func (p *PgUserSecretStore) GetUserSecretEntriesByAppId(userId uuid.UUID, appId uuid.UUID) ([]UserSecretEntry, error) {
	userSecrets := make([]UserSecretEntry, 0)
	qry := infra_db_pg.New(p.DbConn)
	params := &infra_db_pg.GetUserSecretsByAppIdParams{UserID: userId, ApplicationID: appId}
	tokens, err := qry.GetUserSecretsByAppId(context.Background(), *params)
	if err != nil {
		slog.Error("Error retrieving token metadata from database", slog.String("error", err.Error()))
		return userSecrets, err
	}
	for _, entryDb := range tokens {
		var entry UserSecretEntry
		entry.ParseExternalAppSecretMetadataFromAppId(entryDb)
		userSecrets = append(userSecrets, entry)
	}

	return userSecrets, nil
}

func (p *PgUserSecretStore) GetUserSecretEntriesByAppName(userId uuid.UUID, appName string) ([]UserSecretEntry, error) {
	userSecrets := make([]UserSecretEntry, 0)
	qry := infra_db_pg.New(p.DbConn)
	params := &infra_db_pg.GetUserSecretsByAppNameParams{UserID: userId, ApplicationName: appName}
	tokens, err := qry.GetUserSecretsByAppName(context.Background(), *params)
	if err != nil {
		slog.Error("Error retrieving token metadata from database", slog.String("error", err.Error()))
		return userSecrets, err
	}
	for _, entryDb := range tokens {
		var entry UserSecretEntry
		entry.ParseExternalAppSecretMetadataFromAppName(entryDb)
		userSecrets = append(userSecrets, entry)
	}

	return userSecrets, nil
}
