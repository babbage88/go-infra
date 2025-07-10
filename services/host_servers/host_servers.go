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

	// Create host server type mappings if provided
	if len(req.HostServerTypeIDs) > 0 {
		for _, typeID := range req.HostServerTypeIDs {
			_, err = p.db.CreateHostServerTypeMapping(ctx, infra_db_pg.CreateHostServerTypeMappingParams{
				HostServerID:     server.ID,
				HostServerTypeID: typeID,
			})
			if err != nil {
				// Clean up the host server if mapping fails
				_ = p.db.DeleteHostServer(ctx, server.ID)
				return nil, fmt.Errorf("failed to create host server type mapping: %w", err)
			}
		}
	}

	// Create platform type mappings if provided
	if len(req.PlatformTypeIDs) > 0 {
		for _, platformID := range req.PlatformTypeIDs {
			// For platform types, we need to associate with a host server type
			// For now, we'll use the first host server type or create a default mapping
			var hostServerTypeID uuid.UUID
			if len(req.HostServerTypeIDs) > 0 {
				hostServerTypeID = req.HostServerTypeIDs[0]
			} else {
				// Get a default host server type (e.g., "Application Server")
				hostServerType, err := p.db.GetHostServerTypeByName(ctx, "Application Server")
				if err != nil {
					// Clean up the host server if we can't get default type
					_ = p.db.DeleteHostServer(ctx, server.ID)
					return nil, fmt.Errorf("failed to get default host server type: %w", err)
				}
				hostServerTypeID = hostServerType.HostServerTypeID
			}

			_, err = p.db.CreatePlatformTypeMapping(ctx, infra_db_pg.CreatePlatformTypeMappingParams{
				PlatformTypeID:   platformID,
				HostServerID:     server.ID,
				HostServerTypeID: hostServerTypeID,
			})
			if err != nil {
				// Clean up the host server if mapping fails
				_ = p.db.DeleteHostServer(ctx, server.ID)
				return nil, fmt.Errorf("failed to create platform type mapping: %w", err)
			}
		}
	}

	// Get host server types and platform types
	hostServerTypes, platformTypes, err := p.getHostServerTypesAndPlatforms(ctx, server.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get host server types and platforms: %w", err)
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
		HostServerTypes:      hostServerTypes,
		PlatformTypes:        platformTypes,
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

	// Get host server types and platform types
	hostServerTypes, platformTypes, err := p.getHostServerTypesAndPlatforms(ctx, server.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get host server types and platforms: %w", err)
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
		HostServerTypes:      hostServerTypes,
		PlatformTypes:        platformTypes,
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

// CreateHostServerTypeMapping creates a mapping between a host server and a host server type
func (p *HostServerProviderImpl) CreateHostServerTypeMapping(ctx context.Context, hostServerID, hostServerTypeID uuid.UUID) error {
	_, err := p.db.CreateHostServerTypeMapping(ctx, infra_db_pg.CreateHostServerTypeMappingParams{
		HostServerID:     hostServerID,
		HostServerTypeID: hostServerTypeID,
	})
	return err
}

// CreatePlatformTypeMapping creates a mapping between a host server, platform type, and host server type
func (p *HostServerProviderImpl) CreatePlatformTypeMapping(ctx context.Context, hostServerID, platformTypeID, hostServerTypeID uuid.UUID) error {
	_, err := p.db.CreatePlatformTypeMapping(ctx, infra_db_pg.CreatePlatformTypeMappingParams{
		HostServerID:     hostServerID,
		PlatformTypeID:   platformTypeID,
		HostServerTypeID: hostServerTypeID,
	})
	return err
}

// getHostServerTypesAndPlatforms retrieves host server types and platform types for a given host server
func (p *HostServerProviderImpl) getHostServerTypesAndPlatforms(ctx context.Context, hostServerID uuid.UUID) ([]HostServerType, []PlatformType, error) {
	// Get host server types
	hostServerTypeMappings, err := p.db.GetHostServerTypeMappingsByHostId(ctx, hostServerID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get host server type mappings: %w", err)
	}

	hostServerTypes := make([]HostServerType, 0, len(hostServerTypeMappings))
	for _, mapping := range hostServerTypeMappings {
		hostServerTypes = append(hostServerTypes, HostServerType{
			ID:           mapping.HostServerTypeID,
			Name:         mapping.HostServerTypeName,
			LastModified: mapping.LastModified.Time,
		})
	}

	// Get platform types
	platformTypeMappings, err := p.db.GetPlatformTypeMappingsByHostId(ctx, hostServerID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get platform type mappings: %w", err)
	}

	platformTypes := make([]PlatformType, 0, len(platformTypeMappings))
	for _, mapping := range platformTypeMappings {
		platformTypes = append(platformTypes, PlatformType{
			ID:           mapping.PlatformTypeID,
			Name:         mapping.PlatformTypeName,
			LastModified: mapping.LastModified.Time,
		})
	}

	return hostServerTypes, platformTypes, nil
}

// GetAllHostServerTypes retrieves all available host server types
func (p *HostServerProviderImpl) GetAllHostServerTypes(ctx context.Context) ([]HostServerType, error) {
	hostServerTypes, err := p.db.GetAllHostServerTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all host server types: %w", err)
	}

	result := make([]HostServerType, 0, len(hostServerTypes))
	for _, hostServerType := range hostServerTypes {
		result = append(result, HostServerType{
			ID:           hostServerType.HostServerTypeID,
			Name:         hostServerType.Name,
			LastModified: hostServerType.LastModified.Time,
		})
	}

	return result, nil
}

// GetAllPlatformTypes retrieves all available platform types
func (p *HostServerProviderImpl) GetAllPlatformTypes(ctx context.Context) ([]PlatformType, error) {
	platformTypes, err := p.db.GetAllPlatformTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all platform types: %w", err)
	}

	result := make([]PlatformType, 0, len(platformTypes))
	for _, platformType := range platformTypes {
		result = append(result, PlatformType{
			ID:           platformType.PlatformTypeID,
			Name:         platformType.Name,
			LastModified: platformType.LastModified.Time,
		})
	}

	return result, nil
}
