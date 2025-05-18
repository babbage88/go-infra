package user_secrets

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RetrievedUserSecret struct {
	Reader   io.Reader
	Metadata *ExternalApplicationAuthToken
}

type UserSecretProvider interface {
	StoreSecret(plaintextSecret string, userId, appId uuid.UUID) error
	RetrieveSecret(secretId uuid.UUID) (*RetrievedUserSecret, error)
}

// Implementing UserSecretProvider for scenarios where the user supplied secret has only one value that needs to be encrypted
// such as JWT a single bearer token such as a cloudflare token
type PgUserSecretStore struct {
	DbConn *pgxpool.Pool `json:"dbConn"`
}

type PgEncrytpedAuthToken struct {
	UserId        uuid.UUID                      `json:"userId"`
	ApplicationId uuid.UUID                      `json:"applicationId"`
	UserSecret    *EncryptedUserSecretsAES256GCM `json:"userSecret"`
}

func (p *PgUserSecretStore) StoreSecret(plaintextSecret string, userId, appId uuid.UUID) error {
	userCipherText, err := Encrypt(plaintextSecret)
	if err != nil {
		slog.Error("Error encrypting user secret", slog.String("Error", err.Error()))
		return err
	}

	userSecret := PgEncrytpedAuthToken{UserId: userId, ApplicationId: appId, UserSecret: &userCipherText}

	jsonData, err := json.Marshal(userSecret)
	if err != nil {
		slog.Error("Failed to marshal encrypted secret to JSON", slog.String("error", err.Error()))
		return err
	}

	qry := infra_db_pg.New(p.DbConn)
	params := infra_db_pg.InsertExternalAuthTokenParams{UserID: userId, ExternalAppID: appId, Token: jsonData}
	return qry.InsertExternalAuthToken(context.Background(), params)
}

func (p *PgUserSecretStore) RetrieveSecret(secretId uuid.UUID) (*RetrievedUserSecret, error) {
	qry := infra_db_pg.New(p.DbConn)
	record, err := qry.GetExternalAuthTokenById(context.Background(), secretId)
	if err != nil {
		slog.Error("Error retrieving user secret from database", slog.String("error", err.Error()))
		return nil, err
	}

	var encSecret EncryptedUserSecretsAES256GCM
	err = json.Unmarshal(record.Token, &encSecret)
	if err != nil {
		slog.Error("Failed to unmarshal encrypted secret", slog.String("error", err.Error()))
		return nil, err
	}

	plaintext, err := encSecret.Decrypt()
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

	retVal := &RetrievedUserSecret{Metadata: &daoExtSecret, Reader: bytes.NewReader(plaintext)}

	return retVal, nil
}
