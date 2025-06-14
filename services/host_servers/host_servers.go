package host_servers

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// HostServerProviderImpl implements the HostServerProvider interface using PostgreSQL
type HostServerProviderImpl struct {
	db *infra_db_pg.Queries
}

// NewHostServerProvider creates a new HostServerProvider instance
func NewHostServerProvider(db *infra_db_pg.Queries) *HostServerProviderImpl {
	return &HostServerProviderImpl{db: db}
}

// CreateHostServer creates a new host server
func (p *HostServerProviderImpl) CreateHostServer(ctx context.Context, req CreateHostServerRequest) (*HostServer, error) {
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

	return &HostServer{
		ID:               server.ID,
		Hostname:         server.Hostname,
		IPAddress:        server.IpAddress,
		IsContainerHost:  server.IsContainerHost.Bool,
		IsVmHost:         server.IsVmHost.Bool,
		IsVirtualMachine: server.IsVirtualMachine.Bool,
		IsDbHost:         server.IDDbHost.Bool,
		CreatedAt:        server.CreatedAt.Time,
		LastModified:     server.LastModified.Time,
	}, nil
}

// GetHostServer retrieves a host server by ID
func (p *HostServerProviderImpl) GetHostServer(ctx context.Context, id uuid.UUID) (*HostServer, error) {
	server, err := p.db.GetHostServerById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get host server: %w", err)
	}

	return &HostServer{
		ID:               server.ID,
		Hostname:         server.Hostname,
		IPAddress:        server.IpAddress,
		IsContainerHost:  server.IsContainerHost.Bool,
		IsVmHost:         server.IsVmHost.Bool,
		IsVirtualMachine: server.IsVirtualMachine.Bool,
		IsDbHost:         server.IDDbHost.Bool,
		CreatedAt:        server.CreatedAt.Time,
		LastModified:     server.LastModified.Time,
	}, nil
}

// GetHostServerByHostname retrieves a host server by hostname
func (p *HostServerProviderImpl) GetHostServerByHostname(ctx context.Context, hostname string) (*HostServer, error) {
	server, err := p.db.GetHostServerByHostname(ctx, hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to get host server by hostname: %w", err)
	}

	return &HostServer{
		ID:               server.ID,
		Hostname:         server.Hostname,
		IPAddress:        server.IpAddress,
		IsContainerHost:  server.IsContainerHost.Bool,
		IsVmHost:         server.IsVmHost.Bool,
		IsVirtualMachine: server.IsVirtualMachine.Bool,
		IsDbHost:         server.IDDbHost.Bool,
		CreatedAt:        server.CreatedAt.Time,
		LastModified:     server.LastModified.Time,
	}, nil
}

// GetHostServerByIP retrieves a host server by IP address
func (p *HostServerProviderImpl) GetHostServerByIP(ctx context.Context, ip netip.Addr) (*HostServer, error) {
	server, err := p.db.GetHostServerByIP(ctx, ip)
	if err != nil {
		return nil, fmt.Errorf("failed to get host server by IP: %w", err)
	}

	return &HostServer{
		ID:               server.ID,
		Hostname:         server.Hostname,
		IPAddress:        server.IpAddress,
		IsContainerHost:  server.IsContainerHost.Bool,
		IsVmHost:         server.IsVmHost.Bool,
		IsVirtualMachine: server.IsVirtualMachine.Bool,
		IsDbHost:         server.IDDbHost.Bool,
		CreatedAt:        server.CreatedAt.Time,
		LastModified:     server.LastModified.Time,
	}, nil
}

// GetAllHostServers retrieves all host servers
func (p *HostServerProviderImpl) GetAllHostServers(ctx context.Context) ([]HostServer, error) {
	servers, err := p.db.GetAllHostServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all host servers: %w", err)
	}

	result := make([]HostServer, len(servers))
	for i, server := range servers {
		result[i] = HostServer{
			ID:               server.ID,
			Hostname:         server.Hostname,
			IPAddress:        server.IpAddress,
			IsContainerHost:  server.IsContainerHost.Bool,
			IsVmHost:         server.IsVmHost.Bool,
			IsVirtualMachine: server.IsVirtualMachine.Bool,
			IsDbHost:         server.IDDbHost.Bool,
			CreatedAt:        server.CreatedAt.Time,
			LastModified:     server.LastModified.Time,
		}
	}

	return result, nil
}

// UpdateHostServer updates an existing host server
func (p *HostServerProviderImpl) UpdateHostServer(ctx context.Context, id uuid.UUID, req UpdateHostServerRequest) (*HostServer, error) {
	params := infra_db_pg.UpdateHostServerParams{
		ID: id,
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

	server, err := p.db.UpdateHostServer(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update host server: %w", err)
	}

	return &HostServer{
		ID:               server.ID,
		Hostname:         server.Hostname,
		IPAddress:        server.IpAddress,
		IsContainerHost:  server.IsContainerHost.Bool,
		IsVmHost:         server.IsVmHost.Bool,
		IsVirtualMachine: server.IsVirtualMachine.Bool,
		IsDbHost:         server.IDDbHost.Bool,
		CreatedAt:        server.CreatedAt.Time,
		LastModified:     server.LastModified.Time,
	}, nil
}

// DeleteHostServer deletes a host server
func (p *HostServerProviderImpl) DeleteHostServer(ctx context.Context, id uuid.UUID) error {
	err := p.db.DeleteHostServer(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete host server: %w", err)
	}
	return nil
}
