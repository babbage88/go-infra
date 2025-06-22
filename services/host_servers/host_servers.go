package host_servers

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/babbage88/go-infra/api/authapi"
	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// HostServerProviderImpl implements the HostServerProvider interface using PostgreSQL
type HostServerProviderImpl struct {
	db             *infra_db_pg.Queries
	secretProvider user_secrets.UserSecretProvider
}

// NewHostServerProvider creates a new HostServerProvider instance
func NewHostServerProvider(db *infra_db_pg.Queries, secretProvider user_secrets.UserSecretProvider) *HostServerProviderImpl {
	return &HostServerProviderImpl{db: db, secretProvider: secretProvider}
}

// CreateHostServer creates a new host server
func (p *HostServerProviderImpl) CreateHostServer(ctx context.Context, req CreateHostServerRequest) (*HostServer, error) {
	// Validate sudo password token if provided
	if req.SudoPasswordTokenID != nil {
		userId, err := authapi.GetUserIDFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get user ID from context: %w", err)
		}
		secret, err := p.secretProvider.RetrieveSecret(*req.SudoPasswordTokenID)
		if err != nil || secret.ExternalAuthToken.UserID != userId {
			return nil, fmt.Errorf("invalid sudo secret_id provided")
		}
	}

	params := infra_db_pg.CreateHostServerParams{
		Hostname:         req.Hostname,
		IpAddress:        req.IPAddress,
		IsContainerHost:  pgtype.Bool{Bool: req.IsContainerHost, Valid: true},
		IsVmHost:         pgtype.Bool{Bool: req.IsVmHost, Valid: true},
		IsVirtualMachine: pgtype.Bool{Bool: req.IsVirtualMachine, Valid: true},
		IDDbHost:         pgtype.Bool{Bool: req.IsDbHost, Valid: true},
	}

	server, err := p.db.CreateHostServer(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create host server: %w", err)
	}

	userId, err := authapi.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID from context: %w", err)
	}

	// Create SSH key mapping if provided
	if req.SSHKeyID != nil {
		username := ""
		if req.Username != nil {
			username = *req.Username
		}

		var sudoPasswordToken pgtype.UUID
		if req.SudoPasswordTokenID != nil {
			sudoPasswordToken = pgtype.UUID{Bytes: *req.SudoPasswordTokenID, Valid: true}
		}

		_, err = p.db.CreateSSHKeyHostMapping(ctx, infra_db_pg.CreateSSHKeyHostMappingParams{
			SshKeyID:            *req.SSHKeyID,
			HostServerID:        server.ID,
			UserID:              userId,
			HostserverUsername:  username,
			SudoPasswordTokenID: sudoPasswordToken,
		})
		if err != nil {
			// Clean up the host server if mapping fails
			_ = p.db.DeleteHostServer(ctx, server.ID)
			return nil, fmt.Errorf("failed to create SSH key mapping: %w", err)
		}
	}

	return &HostServer{
		ID:                   server.ID,
		Hostname:             server.Hostname,
		IPAddress:            server.IpAddress,
		Username:             req.Username,
		SSHKeyID:             req.SSHKeyID,
		SudoPasswordSecretID: req.SudoPasswordTokenID,
		IsContainerHost:      server.IsContainerHost.Bool,
		IsVmHost:             server.IsVmHost.Bool,
		IsVirtualMachine:     server.IsVirtualMachine.Bool,
		IsDbHost:             server.IDDbHost.Bool,
		CreatedAt:            server.CreatedAt.Time,
		LastModified:         server.LastModified.Time,
	}, nil
}

// GetHostServer retrieves a host server by ID
func (p *HostServerProviderImpl) GetHostServer(ctx context.Context, id uuid.UUID) (*HostServer, error) {
	server, err := p.db.GetHostServerById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get host server: %w", err)
	}

	// Get SSH key mapping if exists
	var username *string
	var sshKeyID *uuid.UUID
	var sudoPasswordTokenID *uuid.UUID
	mappings, err := p.db.GetSSHKeyHostMappingsByHostId(ctx, id)
	if err == nil && len(mappings) > 0 {
		username = &mappings[0].HostserverUsername
		sshKeyID = &mappings[0].SshKeyID
		if mappings[0].SudoPasswordTokenID.Valid {
			tokenID := uuid.UUID(mappings[0].SudoPasswordTokenID.Bytes)
			sudoPasswordTokenID = &tokenID
		}
	}

	return &HostServer{
		ID:                   server.ID,
		Hostname:             server.Hostname,
		IPAddress:            server.IpAddress,
		Username:             username,
		SSHKeyID:             sshKeyID,
		SudoPasswordSecretID: sudoPasswordTokenID,
		IsContainerHost:      server.IsContainerHost.Bool,
		IsVmHost:             server.IsVmHost.Bool,
		IsVirtualMachine:     server.IsVirtualMachine.Bool,
		IsDbHost:             server.IDDbHost.Bool,
		CreatedAt:            server.CreatedAt.Time,
		LastModified:         server.LastModified.Time,
	}, nil
}

// GetHostServerByHostname retrieves a host server by hostname
func (p *HostServerProviderImpl) GetHostServerByHostname(ctx context.Context, hostname string) (*HostServer, error) {
	server, err := p.db.GetHostServerByHostname(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to get host server by hostname: %w", err)
	}

	// Get SSH key mapping if exists
	var username *string
	var sshKeyID *uuid.UUID
	var sudoPasswordTokenID *uuid.UUID
	mappings, err := p.db.GetSSHKeyHostMappingsByHostId(ctx, server.ID)
	if err == nil && len(mappings) > 0 {
		username = &mappings[0].HostserverUsername
		sshKeyID = &mappings[0].SshKeyID
		if mappings[0].SudoPasswordTokenID.Valid {
			tokenID := uuid.UUID(mappings[0].SudoPasswordTokenID.Bytes)
			sudoPasswordTokenID = &tokenID
		}
	}

	return &HostServer{
		ID:                   server.ID,
		Hostname:             server.Hostname,
		IPAddress:            server.IpAddress,
		Username:             username,
		SSHKeyID:             sshKeyID,
		SudoPasswordSecretID: sudoPasswordTokenID,
		IsContainerHost:      server.IsContainerHost.Bool,
		IsVmHost:             server.IsVmHost.Bool,
		IsVirtualMachine:     server.IsVirtualMachine.Bool,
		IsDbHost:             server.IDDbHost.Bool,
		CreatedAt:            server.CreatedAt.Time,
		LastModified:         server.LastModified.Time,
	}, nil
}

// GetHostServerByIP retrieves a host server by IP address
func (p *HostServerProviderImpl) GetHostServerByIP(ctx context.Context, ip netip.Addr) (*HostServer, error) {
	server, err := p.db.GetHostServerByIP(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("failed to get host server by IP: %w", err)
	}

	// Get SSH key mapping if exists
	var username *string
	var sshKeyID *uuid.UUID
	var sudoPasswordTokenID *uuid.UUID
	mappings, err := p.db.GetSSHKeyHostMappingsByHostId(ctx, server.ID)
	if err == nil && len(mappings) > 0 {
		username = &mappings[0].HostserverUsername
		sshKeyID = &mappings[0].SshKeyID
		if mappings[0].SudoPasswordTokenID.Valid {
			tokenID := uuid.UUID(mappings[0].SudoPasswordTokenID.Bytes)
			sudoPasswordTokenID = &tokenID
		}
	}

	return &HostServer{
		ID:                   server.ID,
		Hostname:             server.Hostname,
		IPAddress:            server.IpAddress,
		Username:             username,
		SSHKeyID:             sshKeyID,
		SudoPasswordSecretID: sudoPasswordTokenID,
		IsContainerHost:      server.IsContainerHost.Bool,
		IsVmHost:             server.IsVmHost.Bool,
		IsVirtualMachine:     server.IsVirtualMachine.Bool,
		IsDbHost:             server.IDDbHost.Bool,
		CreatedAt:            server.CreatedAt.Time,
		LastModified:         server.LastModified.Time,
	}, nil
}

// GetAllHostServers retrieves all host servers
func (p *HostServerProviderImpl) GetAllHostServers(ctx context.Context) ([]HostServer, error) {
	servers, err := p.db.GetAllHostServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all host servers: %w", err)
	}

	result := make([]HostServer, 0, len(servers))
	for _, server := range servers {
		// Get SSH key mapping if exists
		var username *string
		var sshKeyID *uuid.UUID
		var sudoPasswordTokenID *uuid.UUID
		mappings, err := p.db.GetSSHKeyHostMappingsByHostId(ctx, server.ID)
		if err == nil && len(mappings) > 0 {
			username = &mappings[0].HostserverUsername
			sshKeyID = &mappings[0].SshKeyID
			if mappings[0].SudoPasswordTokenID.Valid {
				tokenID := uuid.UUID(mappings[0].SudoPasswordTokenID.Bytes)
				sudoPasswordTokenID = &tokenID
			}
		}

		result = append(result, HostServer{
			ID:                   server.ID,
			Hostname:             server.Hostname,
			IPAddress:            server.IpAddress,
			Username:             username,
			SSHKeyID:             sshKeyID,
			SudoPasswordSecretID: sudoPasswordTokenID,
			IsContainerHost:      server.IsContainerHost.Bool,
			IsVmHost:             server.IsVmHost.Bool,
			IsVirtualMachine:     server.IsVirtualMachine.Bool,
			IsDbHost:             server.IDDbHost.Bool,
			CreatedAt:            server.CreatedAt.Time,
			LastModified:         server.LastModified.Time,
		})
	}

	return result, nil
}

// UpdateHostServer updates an existing host server
func (p *HostServerProviderImpl) UpdateHostServer(ctx context.Context, id uuid.UUID, req UpdateHostServerRequest) (*HostServer, error) {
	// Validate sudo password token if provided
	if req.SudoPasswordTokenID != nil {
		userId, err := authapi.GetUserIDFromContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get user ID from context: %w", err)
		}
		secret, err := p.secretProvider.RetrieveSecret(*req.SudoPasswordTokenID)
		if err != nil || secret.ExternalAuthToken.UserID != userId {
			return nil, fmt.Errorf("invalid sudo secret_id provided")
		}
	}

	// Get current server to merge with updates
	current, err := p.GetHostServer(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get user ID from context
	userId, err := authapi.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID from context: %w", err)
	}

	params := infra_db_pg.UpdateHostServerParams{
		ID:               id,
		Hostname:         current.Hostname,
		IpAddress:        current.IPAddress,
		IsContainerHost:  pgtype.Bool{Bool: current.IsContainerHost, Valid: true},
		IsVmHost:         pgtype.Bool{Bool: current.IsVmHost, Valid: true},
		IsVirtualMachine: pgtype.Bool{Bool: current.IsVirtualMachine, Valid: true},
		IDDbHost:         pgtype.Bool{Bool: current.IsDbHost, Valid: true},
	}

	if req.Hostname != nil {
		params.Hostname = *req.Hostname
	}
	if req.IPAddress != nil {
		params.IpAddress = *req.IPAddress
	}
	if req.IsContainerHost != nil {
		params.IsContainerHost = pgtype.Bool{Bool: *req.IsContainerHost, Valid: true}
	}
	if req.IsVmHost != nil {
		params.IsVmHost = pgtype.Bool{Bool: *req.IsVmHost, Valid: true}
	}
	if req.IsVirtualMachine != nil {
		params.IsVirtualMachine = pgtype.Bool{Bool: *req.IsVirtualMachine, Valid: true}
	}
	if req.IsDbHost != nil {
		params.IDDbHost = pgtype.Bool{Bool: *req.IsDbHost, Valid: true}
	}

	// Update host server
	_, err = p.db.UpdateHostServer(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update host server: %w", err)
	}

	// Update SSH key mapping if provided
	if req.SSHKeyID != nil || req.Username != nil {
		mappings, err := p.db.GetSSHKeyHostMappingsByHostId(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get SSH key mapping: %w", err)
		}

		username := ""
		if current.Username != nil {
			username = *current.Username
		}
		if req.Username != nil {
			username = *req.Username
		}

		if len(mappings) > 0 {
			// Update existing mapping
			_, err = p.db.UpdateSSHKeyHostMapping(ctx, infra_db_pg.UpdateSSHKeyHostMappingParams{
				ID:                 mappings[0].MappingID,
				HostserverUsername: username,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to update SSH key mapping: %w", err)
			}
		} else if req.SSHKeyID != nil {
			// Create new mapping
			var sudoPasswordToken pgtype.UUID
			if req.SudoPasswordTokenID != nil {
				sudoPasswordToken = pgtype.UUID{Bytes: *req.SudoPasswordTokenID, Valid: true}
			}
			_, err = p.db.CreateSSHKeyHostMapping(ctx, infra_db_pg.CreateSSHKeyHostMappingParams{
				SshKeyID:            *req.SSHKeyID,
				HostServerID:        id,
				UserID:              userId,
				HostserverUsername:  username,
				SudoPasswordTokenID: sudoPasswordToken,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create SSH key mapping: %w", err)
			}
		}
	}

	// Update sudo password token if provided
	if req.SudoPasswordTokenID != nil {
		// Delete existing tokens
		tokens, err := p.db.GetExternalAuthTokensByUserIdAndAppId(ctx, infra_db_pg.GetExternalAuthTokensByUserIdAndAppIdParams{
			UserID:        userId,
			ExternalAppID: id,
		})
		if err == nil {
			for _, token := range tokens {
				_ = p.db.DeleteExternalAuthTokenById(ctx, token.ID)
			}
		}

		// Create new token
		_, err = p.secretProvider.StoreSecret(uuid.New().String(), userId, id, time.Now().Add(24*time.Hour))
		if err != nil {
			return nil, fmt.Errorf("failed to update sudo password token: %w", err)
		}
	}

	return p.GetHostServer(ctx, id)
}

// DeleteHostServer deletes a host server
func (p *HostServerProviderImpl) DeleteHostServer(ctx context.Context, id uuid.UUID) error {
	err := p.db.DeleteHostServer(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete host server: %w", err)
	}
	return nil
}
