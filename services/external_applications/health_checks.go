package external_applications

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExternalApplicationsHealthCheck struct {
	DbConn *pgxpool.Pool
}

// HealthCheck performs a basic health check on the external applications service
func (eahc *ExternalApplicationsHealthCheck) HealthCheck() error {
	slog.Info("Performing external applications service health check")

	queries := infra_db_pg.New(eahc.DbConn)

	// Try to get all external applications as a health check
	_, err := queries.GetAllExternalApps(context.Background())
	if err != nil {
		slog.Error("External applications health check failed", slog.String("error", err.Error()))
		return fmt.Errorf("external applications service health check failed: %w", err)
	}

	slog.Info("External applications service health check passed")
	return nil
}

// DatabaseHealthCheck performs a database-specific health check
func (eahc *ExternalApplicationsHealthCheck) DatabaseHealthCheck() error {
	slog.Info("Performing external applications database health check")

	queries := infra_db_pg.New(eahc.DbConn)

	// Use the general database health check
	_, err := queries.DbHealthCheckRead(context.Background())
	if err != nil {
		slog.Error("External applications database health check failed", slog.String("error", err.Error()))
		return fmt.Errorf("external applications database health check failed: %w", err)
	}

	slog.Info("External applications database health check passed")
	return nil
}
