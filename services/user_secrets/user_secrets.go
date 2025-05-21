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
	Reader   io.Reader
	Metadata *ExternalApplicationAuthToken
}

type UserSecretProvider interface {
	StoreSecret(plaintextSecret string, userId, appId uuid.UUID, expiry time.Time) error
	RetrieveSecret(secretId uuid.UUID) (*RetrievedUserSecret, error)
	DeleteSecret(secretId uuid.UUID) error
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

func (p *PgUserSecretStore) StoreSecret(plaintextSecret string, userId, appId uuid.UUID, expiry time.Time) error {
	userCipherText, err := Encrypt(plaintextSecret)
	if err != nil {
		slog.Error("Error encrypting user secret", slog.String("Error", err.Error()))
		return err
	}

	userSecret := PgEncrytpedAuthToken{
		UserId:        userId,
		ApplicationId: appId,
		UserSecret:    &userCipherText,
	}

	jsonData, err := json.Marshal(userSecret)
	if err != nil {
		slog.Error("Failed to marshal encrypted secret to JSON", slog.String("error", err.Error()))
		return err
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
	return qry.InsertExternalAuthToken(context.Background(), params)
}

func (p *PgUserSecretStore) RetrieveSecret(secretId uuid.UUID) (*RetrievedUserSecret, error) {
	qry := infra_db_pg.New(p.DbConn)
	record, err := qry.GetExternalAuthTokenById(context.Background(), secretId)
	if err != nil {
		slog.Error("Error retrieving user secret from database", slog.String("error", err.Error()))
		return nil, err
	}

	var stored PgEncrytpedAuthToken
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
		Reader:   bytes.NewReader(plaintext),
		Metadata: &daoExtSecret,
	}, nil
}

func (p *PgUserSecretStore) DeleteSecret(secretId uuid.UUID) error {
	qry := infra_db_pg.New(p.DbConn)
	result := qry.DeleteExternalAuthTokenById(context.Background(), secretId)
	return result
}
