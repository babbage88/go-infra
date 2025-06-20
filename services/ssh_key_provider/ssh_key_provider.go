package ssh_key_provider

import (
	"context"
	"log/slog"
	"time"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgSshKeySecretStore struct {
	DbConn         *pgxpool.Pool
	SecretProvider user_secrets.UserSecretProvider
}

func NewPgSshKeySecretStore(dbConn *pgxpool.Pool, secretProvider user_secrets.UserSecretProvider) *PgSshKeySecretStore {
	return &PgSshKeySecretStore{
		DbConn:         dbConn,
		SecretProvider: secretProvider,
	}
}

func (p *PgSshKeySecretStore) StoreSshKeySecret(plaintextSecret string, userId, appId uuid.UUID, expiry time.Time) error {
	return p.SecretProvider.StoreSecret(plaintextSecret, userId, appId, expiry)
}

func (p *PgSshKeySecretStore) CreateSshKey(sshKey *NewSshKeyRequest) NewSshKeyResult {
	// Start a transaction
	tx, err := p.DbConn.Begin(context.Background())
	if err != nil {
		slog.Error("Failed to begin transaction", slog.String("error", err.Error()))
		return NewSshKeyResult{Error: err}
	}
	defer tx.Rollback(context.Background())

	qry := infra_db_pg.New(tx)

	// Get the SSH key type ID
	keyType, err := qry.GetSSHKeyTypeByName(context.Background(), sshKey.KeyType)
	if err != nil {
		slog.Error("Failed to get SSH key type", slog.String("error", err.Error()))
		return NewSshKeyResult{Error: err}
	}

	// Store the private key as a secret
	// Use a default expiry of 1 year if not specified
	expiry := time.Now().AddDate(1, 0, 0)

	// Get the external app ID for SSH keys
	sshAppId, err := qry.GetExternalAppIdByName(context.Background(), "ssh_key")
	if err != nil {
		slog.Error("Failed to get SSH app ID", slog.String("error", err.Error()))
		return NewSshKeyResult{Error: err}
	}

	// Store the private key as a secret
	err = p.StoreSshKeySecret(sshKey.PrivateKey, sshKey.UserID, sshAppId, expiry)
	if err != nil {
		slog.Error("Failed to store SSH key secret", slog.String("error", err.Error()))
		return NewSshKeyResult{Error: err}
	}

	// Get the latest token ID for the stored secret
	token, err := qry.GetLatestExternalAuthToken(context.Background(), infra_db_pg.GetLatestExternalAuthTokenParams{
		UserID:        sshKey.UserID,
		ExternalAppID: sshAppId,
	})
	if err != nil {
		slog.Error("Failed to get stored secret token", slog.String("error", err.Error()))
		return NewSshKeyResult{Error: err}
	}

	// Create the SSH key record
	sshKeyRecord, err := qry.CreateSSHKey(context.Background(), infra_db_pg.CreateSSHKeyParams{
		Name:         sshKey.Name,
		Description:  pgtype.Text{String: sshKey.Description, Valid: true},
		PrivSecretID: pgtype.UUID{Bytes: token.ID, Valid: true},
		PublicKey:    sshKey.PublicKey,
		KeyTypeID:    keyType.ID,
		OwnerUserID:  sshKey.UserID,
	})
	if err != nil {
		slog.Error("Failed to create SSH key record", slog.String("error", err.Error()))
		return NewSshKeyResult{Error: err}
	}

	// If a host server was specified, create the mapping
	if sshKey.HostServerId != uuid.Nil {
		// Get the default username for the host server
		hostServer, err := qry.GetHostServerById(context.Background(), sshKey.HostServerId)
		if err != nil {
			slog.Error("Failed to get host server", slog.String("error", err.Error()))
			return NewSshKeyResult{Error: err}
		}

		// Create the mapping with the hostname as the default username
		_, err = qry.CreateSSHKeyHostMapping(context.Background(), infra_db_pg.CreateSSHKeyHostMappingParams{
			SshKeyID:           sshKeyRecord.ID,
			HostServerID:       sshKey.HostServerId,
			UserID:             sshKey.UserID,
			HostserverUsername: hostServer.Hostname,
		})
		if err != nil {
			slog.Error("Failed to create SSH key host mapping", slog.String("error", err.Error()))
			return NewSshKeyResult{Error: err}
		}
	}

	// Commit the transaction
	if err := tx.Commit(context.Background()); err != nil {
		slog.Error("Failed to commit transaction", slog.String("error", err.Error()))
		return NewSshKeyResult{Error: err}
	}

	return NewSshKeyResult{
		SshKeyId:        sshKeyRecord.ID,
		PrivKeySecretId: token.ID,
		UserId:          sshKey.UserID,
		Error:           nil,
	}
}

func (p *PgSshKeySecretStore) DeleteSShKeyAndSecret(sshKeyId uuid.UUID) error {
	// Start a transaction
	tx, err := p.DbConn.Begin(context.Background())
	if err != nil {
		slog.Error("Failed to begin transaction", slog.String("error", err.Error()))
		return err
	}
	defer tx.Rollback(context.Background())

	qry := infra_db_pg.New(tx)

	// First, get the SSH key to retrieve the secret ID
	sshKey, err := qry.GetSSHKeyById(context.Background(), sshKeyId)
	if err != nil {
		slog.Error("Failed to get SSH key", slog.String("error", err.Error()))
		return err
	}

	// Delete SSH key host mappings first (foreign key constraint)
	err = qry.DeleteSSHKeyHostMappingsBySshKeyId(context.Background(), sshKeyId)
	if err != nil {
		slog.Error("Failed to delete SSH key host mappings", slog.String("error", err.Error()))
		return err
	}

	// Delete the SSH key record
	err = qry.DeleteSSHKey(context.Background(), sshKeyId)
	if err != nil {
		slog.Error("Failed to delete SSH key", slog.String("error", err.Error()))
		return err
	}

	// Delete the associated secret if it exists
	if sshKey.PrivSecretID.Valid {
		err = p.SecretProvider.DeleteSecret(sshKey.PrivSecretID.Bytes)
		if err != nil {
			slog.Error("Failed to delete SSH key secret", slog.String("error", err.Error()))
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit(context.Background()); err != nil {
		slog.Error("Failed to commit transaction", slog.String("error", err.Error()))
		return err
	}

	return nil
}

// SSH Key Host Mapping CRUD operations

func (p *PgSshKeySecretStore) CreateSshKeyHostMapping(mapping *CreateSshKeyHostMappingRequest) CreateSshKeyHostMappingResult {
	// Start a transaction
	tx, err := p.DbConn.Begin(context.Background())
	if err != nil {
		slog.Error("Failed to begin transaction", slog.String("error", err.Error()))
		return CreateSshKeyHostMappingResult{Error: err}
	}
	defer tx.Rollback(context.Background())

	qry := infra_db_pg.New(tx)

	// Create the SSH key host mapping
	sshKeyHostMapping, err := qry.CreateSSHKeyHostMapping(context.Background(), infra_db_pg.CreateSSHKeyHostMappingParams{
		SshKeyID:           mapping.SshKeyID,
		HostServerID:       mapping.HostServerID,
		UserID:             mapping.UserID,
		HostserverUsername: mapping.HostserverUsername,
	})
	if err != nil {
		slog.Error("Failed to create SSH key host mapping", slog.String("error", err.Error()))
		return CreateSshKeyHostMappingResult{Error: err}
	}

	// Commit the transaction
	if err := tx.Commit(context.Background()); err != nil {
		slog.Error("Failed to commit transaction", slog.String("error", err.Error()))
		return CreateSshKeyHostMappingResult{Error: err}
	}

	return CreateSshKeyHostMappingResult{
		ID:                 sshKeyHostMapping.ID,
		SshKeyID:           sshKeyHostMapping.SshKeyID,
		HostServerID:       sshKeyHostMapping.HostServerID,
		UserID:             sshKeyHostMapping.UserID,
		HostserverUsername: sshKeyHostMapping.HostserverUsername,
		CreatedAt:          sshKeyHostMapping.CreatedAt.Time,
		LastModified:       sshKeyHostMapping.LastModified.Time,
		Error:              nil,
	}
}

func (p *PgSshKeySecretStore) GetSshKeyHostMappingById(id uuid.UUID) (*CreateSshKeyHostMappingResult, error) {
	qry := infra_db_pg.New(p.DbConn)

	// Get the SSH key host mapping
	sshKeyHostMapping, err := qry.GetSSHKeyHostMappingById(context.Background(), id)
	if err != nil {
		slog.Error("Failed to get SSH key host mapping", slog.String("error", err.Error()))
		return nil, err
	}

	return &CreateSshKeyHostMappingResult{
		ID:                 sshKeyHostMapping.ID,
		SshKeyID:           sshKeyHostMapping.SshKeyID,
		HostServerID:       sshKeyHostMapping.HostServerID,
		UserID:             sshKeyHostMapping.UserID,
		HostserverUsername: sshKeyHostMapping.HostserverUsername,
		CreatedAt:          sshKeyHostMapping.CreatedAt.Time,
		LastModified:       sshKeyHostMapping.LastModified.Time,
		Error:              nil,
	}, nil
}

func (p *PgSshKeySecretStore) GetSshKeyHostMappingsByUserId(userId uuid.UUID) ([]CreateSshKeyHostMappingResult, error) {
	qry := infra_db_pg.New(p.DbConn)

	// Get the SSH key host mappings by user ID
	userSshKeyMappings, err := qry.GetSSHKeyHostMappingsByUserId(context.Background(), userId)
	if err != nil {
		slog.Error("Failed to get SSH key host mappings by user ID", slog.String("error", err.Error()))
		return nil, err
	}

	result := make([]CreateSshKeyHostMappingResult, 0, len(userSshKeyMappings))
	for _, mapping := range userSshKeyMappings {
		result = append(result, CreateSshKeyHostMappingResult{
			SshKeyID:           mapping.SshKeyID,
			HostServerID:       mapping.HostServerID,
			UserID:             mapping.UserID,
			HostserverUsername: mapping.HostserverUsername,
			// Note: UserSshKeyMapping doesn't have ID, CreatedAt, LastModified fields
			// These would need to be fetched separately if needed
		})
	}

	return result, nil
}

func (p *PgSshKeySecretStore) GetSshKeyHostMappingsByHostId(hostId uuid.UUID) ([]CreateSshKeyHostMappingResult, error) {
	qry := infra_db_pg.New(p.DbConn)

	// Get the SSH key host mappings by host ID
	userSshKeyMappings, err := qry.GetSSHKeyHostMappingsByHostId(context.Background(), hostId)
	if err != nil {
		slog.Error("Failed to get SSH key host mappings by host ID", slog.String("error", err.Error()))
		return nil, err
	}

	result := make([]CreateSshKeyHostMappingResult, 0, len(userSshKeyMappings))
	for _, mapping := range userSshKeyMappings {
		result = append(result, CreateSshKeyHostMappingResult{
			SshKeyID:           mapping.SshKeyID,
			HostServerID:       mapping.HostServerID,
			UserID:             mapping.UserID,
			HostserverUsername: mapping.HostserverUsername,
			// Note: UserSshKeyMapping doesn't have ID, CreatedAt, LastModified fields
			// These would need to be fetched separately if needed
		})
	}

	return result, nil
}

func (p *PgSshKeySecretStore) GetSshKeyHostMappingsByKeyId(keyId uuid.UUID) ([]CreateSshKeyHostMappingResult, error) {
	qry := infra_db_pg.New(p.DbConn)

	// Get the SSH key host mappings by key ID
	userSshKeyMappings, err := qry.GetSSHKeyHostMappingsByKeyId(context.Background(), keyId)
	if err != nil {
		slog.Error("Failed to get SSH key host mappings by key ID", slog.String("error", err.Error()))
		return nil, err
	}

	result := make([]CreateSshKeyHostMappingResult, 0, len(userSshKeyMappings))
	for _, mapping := range userSshKeyMappings {
		result = append(result, CreateSshKeyHostMappingResult{
			SshKeyID:           mapping.SshKeyID,
			HostServerID:       mapping.HostServerID,
			UserID:             mapping.UserID,
			HostserverUsername: mapping.HostserverUsername,
			// Note: UserSshKeyMapping doesn't have ID, CreatedAt, LastModified fields
			// These would need to be fetched separately if needed
		})
	}

	return result, nil
}

func (p *PgSshKeySecretStore) UpdateSshKeyHostMapping(mapping *UpdateSshKeyHostMappingRequest) UpdateSshKeyHostMappingResult {
	// Start a transaction
	tx, err := p.DbConn.Begin(context.Background())
	if err != nil {
		slog.Error("Failed to begin transaction", slog.String("error", err.Error()))
		return UpdateSshKeyHostMappingResult{Error: err}
	}
	defer tx.Rollback(context.Background())

	qry := infra_db_pg.New(tx)

	// Update the SSH key host mapping
	sshKeyHostMapping, err := qry.UpdateSSHKeyHostMapping(context.Background(), infra_db_pg.UpdateSSHKeyHostMappingParams{
		ID:                 mapping.ID,
		HostserverUsername: mapping.HostserverUsername,
	})
	if err != nil {
		slog.Error("Failed to update SSH key host mapping", slog.String("error", err.Error()))
		return UpdateSshKeyHostMappingResult{Error: err}
	}

	// Commit the transaction
	if err := tx.Commit(context.Background()); err != nil {
		slog.Error("Failed to commit transaction", slog.String("error", err.Error()))
		return UpdateSshKeyHostMappingResult{Error: err}
	}

	return UpdateSshKeyHostMappingResult{
		ID:                 sshKeyHostMapping.ID,
		SshKeyID:           sshKeyHostMapping.SshKeyID,
		HostServerID:       sshKeyHostMapping.HostServerID,
		UserID:             sshKeyHostMapping.UserID,
		HostserverUsername: sshKeyHostMapping.HostserverUsername,
		CreatedAt:          sshKeyHostMapping.CreatedAt.Time,
		LastModified:       sshKeyHostMapping.LastModified.Time,
		Error:              nil,
	}
}

func (p *PgSshKeySecretStore) DeleteSshKeyHostMapping(id uuid.UUID) error {
	// Start a transaction
	tx, err := p.DbConn.Begin(context.Background())
	if err != nil {
		slog.Error("Failed to begin transaction", slog.String("error", err.Error()))
		return err
	}
	defer tx.Rollback(context.Background())

	qry := infra_db_pg.New(tx)

	// Delete the SSH key host mapping
	err = qry.DeleteSSHKeyHostMapping(context.Background(), id)
	if err != nil {
		slog.Error("Failed to delete SSH key host mapping", slog.String("error", err.Error()))
		return err
	}

	// Commit the transaction
	if err := tx.Commit(context.Background()); err != nil {
		slog.Error("Failed to commit transaction", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (p *PgSshKeySecretStore) DeleteSshKeyHostMappingsBySshKeyId(sshKeyId uuid.UUID) error {
	// Start a transaction
	tx, err := p.DbConn.Begin(context.Background())
	if err != nil {
		slog.Error("Failed to begin transaction", slog.String("error", err.Error()))
		return err
	}
	defer tx.Rollback(context.Background())

	qry := infra_db_pg.New(tx)

	// Delete all SSH key host mappings for the given SSH key ID
	err = qry.DeleteSSHKeyHostMappingsBySshKeyId(context.Background(), sshKeyId)
	if err != nil {
		slog.Error("Failed to delete SSH key host mappings by SSH key ID", slog.String("error", err.Error()))
		return err
	}

	// Commit the transaction
	if err := tx.Commit(context.Background()); err != nil {
		slog.Error("Failed to commit transaction", slog.String("error", err.Error()))
		return err
	}

	return nil
}
