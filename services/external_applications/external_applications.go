package external_applications

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExternalApplicationsService struct {
	DbConn *pgxpool.Pool
}

type ExternalApplications interface {
	CreateExternalApplication(req CreateExternalApplicationRequest) (*ExternalApplicationDao, error)
	GetExternalApplicationById(id uuid.UUID) (*ExternalApplicationDao, error)
	GetExternalApplicationByName(name string) (*ExternalApplicationDao, error)
	GetAllExternalApplications() ([]ExternalApplicationDao, error)
	UpdateExternalApplication(id uuid.UUID, req UpdateExternalApplicationRequest) (*ExternalApplicationDao, error)
	DeleteExternalApplicationById(id uuid.UUID) error
	DeleteExternalApplicationByName(name string) error
	GetExternalApplicationIdByName(name string) (uuid.UUID, error)
	GetExternalApplicationNameById(id uuid.UUID) (string, error)
}

// CreateExternalApplication creates a new external application
func (eas *ExternalApplicationsService) CreateExternalApplication(req CreateExternalApplicationRequest) (*ExternalApplicationDao, error) {
	slog.Info("Creating external application", slog.String("name", req.Name))

	// Generate a new UUID for the application
	appId := uuid.New()

	// Set up parameters for the new external application
	params := infra_db_pg.CreateExternalApplicationParams{
		ID:   appId,
		Name: req.Name,
	}

	// Set optional fields if provided
	if req.EndpointUrl != "" {
		params.EndpointUrl = pgtype.Text{String: req.EndpointUrl, Valid: true}
	}
	if req.AppDescription != "" {
		params.AppDescription = pgtype.Text{String: req.AppDescription, Valid: true}
	}

	queries := infra_db_pg.New(eas.DbConn)
	dbApp, err := queries.CreateExternalApplication(context.Background(), params)
	if err != nil {
		slog.Error("Error creating external application",
			slog.String("name", req.Name),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create external application: %w", err)
	}

	// Parse the database result to DAO
	var appDao ExternalApplicationDao
	appDao.ParseExternalApplicationFromDb(dbApp)

	return &appDao, nil
}

// GetExternalApplicationById retrieves an external application by its ID
func (eas *ExternalApplicationsService) GetExternalApplicationById(id uuid.UUID) (*ExternalApplicationDao, error) {
	slog.Info("Getting external application by ID", slog.String("id", id.String()))

	queries := infra_db_pg.New(eas.DbConn)

	dbApp, err := queries.GetExternalApplicationById(context.Background(), id)
	if err != nil {
		slog.Error("Error getting external application by ID",
			slog.String("id", id.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("external application not found with ID %s: %w", id.String(), err)
	}

	// Parse the database result to DAO
	var appDao ExternalApplicationDao
	appDao.ParseExternalApplicationFromDb(dbApp)

	return &appDao, nil
}

// GetExternalApplicationByName retrieves an external application by its name
func (eas *ExternalApplicationsService) GetExternalApplicationByName(name string) (*ExternalApplicationDao, error) {
	slog.Info("Getting external application by name", slog.String("name", name))

	queries := infra_db_pg.New(eas.DbConn)

	dbApp, err := queries.GetExternalApplicationByName(context.Background(), name)
	if err != nil {
		slog.Error("Error getting external application by name",
			slog.String("name", name),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("external application not found with name %s: %w", name, err)
	}

	// Parse the database result to DAO
	var appDao ExternalApplicationDao
	appDao.ParseExternalApplicationFromDb(dbApp)

	return &appDao, nil
}

// GetAllExternalApplications retrieves all external applications
func (eas *ExternalApplicationsService) GetAllExternalApplications() ([]ExternalApplicationDao, error) {
	slog.Info("Getting all external applications")

	queries := infra_db_pg.New(eas.DbConn)
	rows, err := queries.GetAllExternalApps(context.Background())
	if err != nil {
		slog.Error("Error getting all external applications", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get external applications: %w", err)
	}

	// Map rows to ExternalApplicationDao
	appDaos := make([]ExternalApplicationDao, len(rows))
	for i, row := range rows {
		appDaos[i].ParseExternalApplicationFromGetAllRow(row)
	}

	return appDaos, nil
}

// UpdateExternalApplication updates an existing external application
func (eas *ExternalApplicationsService) UpdateExternalApplication(id uuid.UUID, req UpdateExternalApplicationRequest) (*ExternalApplicationDao, error) {
	slog.Info("Updating external application", slog.String("id", id.String()))

	queries := infra_db_pg.New(eas.DbConn)

	// Set up parameters for the update
	params := infra_db_pg.UpdateExternalApplicationParams{
		ID: id,
	}

	// Only set fields that are provided in the request
	if req.Name != "" {
		params.Name = req.Name
	}
	if req.EndpointUrl != "" {
		params.EndpointUrl = pgtype.Text{String: req.EndpointUrl, Valid: true}
	}
	if req.AppDescription != "" {
		params.AppDescription = pgtype.Text{String: req.AppDescription, Valid: true}
	}

	dbApp, err := queries.UpdateExternalApplication(context.Background(), params)
	if err != nil {
		slog.Error("Error updating external application",
			slog.String("id", id.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to update external application: %w", err)
	}

	// Parse the database result to DAO
	var appDao ExternalApplicationDao
	appDao.ParseExternalApplicationFromDb(dbApp)

	return &appDao, nil
}

// DeleteExternalApplicationById deletes an external application by its ID
func (eas *ExternalApplicationsService) DeleteExternalApplicationById(id uuid.UUID) error {
	slog.Info("Deleting external application by ID", slog.String("id", id.String()))

	queries := infra_db_pg.New(eas.DbConn)
	err := queries.DeleteExternalApplicationById(context.Background(), id)
	if err != nil {
		slog.Error("Error deleting external application by ID",
			slog.String("id", id.String()),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to delete external application: %w", err)
	}

	return nil
}

// DeleteExternalApplicationByName deletes an external application by its name
func (eas *ExternalApplicationsService) DeleteExternalApplicationByName(name string) error {
	slog.Info("Deleting external application by name", slog.String("name", name))

	queries := infra_db_pg.New(eas.DbConn)
	err := queries.DeleteExternalApplicationByName(context.Background(), name)
	if err != nil {
		slog.Error("Error deleting external application by name",
			slog.String("name", name),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to delete external application: %w", err)
	}

	return nil
}

// GetExternalApplicationIdByName retrieves the ID of an external application by its name
func (eas *ExternalApplicationsService) GetExternalApplicationIdByName(name string) (uuid.UUID, error) {
	slog.Info("Getting external application ID by name", slog.String("name", name))

	queries := infra_db_pg.New(eas.DbConn)
	id, err := queries.GetExternalAppIdByName(context.Background(), name)
	if err != nil {
		slog.Error("Error getting external application ID by name",
			slog.String("name", name),
			slog.String("error", err.Error()))
		return uuid.Nil, fmt.Errorf("external application not found with name %s: %w", name, err)
	}

	return id, nil
}

// GetExternalApplicationNameById retrieves the name of an external application by its ID
func (eas *ExternalApplicationsService) GetExternalApplicationNameById(id uuid.UUID) (string, error) {
	slog.Info("Getting external application name by ID", slog.String("id", id.String()))

	queries := infra_db_pg.New(eas.DbConn)
	name, err := queries.GetExternalAppNameById(context.Background(), id)
	if err != nil {
		slog.Error("Error getting external application name by ID",
			slog.String("id", id.String()),
			slog.String("error", err.Error()))
		return "", fmt.Errorf("external application not found with ID %s: %w", id.String(), err)
	}

	return name, nil
}
