package user_applications

import (
	"context"
	"fmt"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserApplicationsService struct {
	DbConn *pgxpool.Pool
}

type UserApplications interface {
	CreateUserApplication(req CreateUserApplicationRequest) (*UserApplicationDao, error)
	GetUserApplicationById(id uuid.UUID) (*UserApplicationDao, error)
	GetUserApplicationByName(name string) (*UserApplicationDao, error)
	GetAllUserApplications() ([]UserApplicationDao, error)
	UpdateUserApplication(id uuid.UUID, req UpdateUserApplicationRequest) (*UserApplicationDao, error)
	DeleteUserApplicationById(id uuid.UUID) error
	DeleteUserApplicationByName(name string) error
}

func normalizeCreateRequest(req CreateUserApplicationRequest) CreateUserApplicationRequest {
	if req.SourceKind == "" {
		req.SourceKind = "manual"
	}
	if req.Registerable == nil {
		req.Registerable = boolPtr(true)
	}
	return req
}

func (svc *UserApplicationsService) CreateUserApplication(req CreateUserApplicationRequest) (*UserApplicationDao, error) {
	req = normalizeCreateRequest(req)
	queries := infra_db_pg.New(svc.DbConn)
	ctx := context.Background()

	deployConfig, err := marshalOptionalJSON(req.DeployConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal deploy config: %w", err)
	}
	buildConfig, err := marshalOptionalJSON(req.BuildConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal build config: %w", err)
	}

	tx, err := svc.DbConn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	qtx := queries.WithTx(tx)

	dbApp, err := qtx.CreateUserApplication(ctx, infra_db_pg.CreateUserApplicationParams{
		Name:           req.Name,
		Description:    pgtype.Text{String: req.Description, Valid: req.Description != ""},
		RepositoryUrl:  req.RepositoryUrl,
		ManifestPath:   pgtype.Text{String: req.ManifestPath, Valid: req.ManifestPath != ""},
		SourceKind:     req.SourceKind,
		ModuleName:     pgtype.Text{String: req.ModuleName, Valid: req.ModuleName != ""},
		PackageName:    pgtype.Text{String: req.PackageName, Valid: req.PackageName != ""},
		PackageManager: pgtype.Text{String: req.PackageManager, Valid: req.PackageManager != ""},
		DeployKind:     req.DeployKind,
		Registerable:   boolValue(req.Registerable),
		DeployConfig:   deployConfig,
		BuildConfig:    buildConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("create user application: %w", err)
	}

	deps, err := createInfraDependencies(ctx, qtx, dbApp.ID, req.InfraDependencies)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	app := parseUserApplication(dbApp, deps)
	return &app, nil
}

func (svc *UserApplicationsService) GetUserApplicationById(id uuid.UUID) (*UserApplicationDao, error) {
	queries := infra_db_pg.New(svc.DbConn)
	ctx := context.Background()
	dbApp, err := queries.GetUserApplicationById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user application by id: %w", err)
	}
	deps, err := getInfraDependencies(ctx, queries, id)
	if err != nil {
		return nil, err
	}
	app := parseUserApplication(dbApp, deps)
	return &app, nil
}

func (svc *UserApplicationsService) GetUserApplicationByName(name string) (*UserApplicationDao, error) {
	queries := infra_db_pg.New(svc.DbConn)
	ctx := context.Background()
	dbApp, err := queries.GetUserApplicationByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get user application by name: %w", err)
	}
	deps, err := getInfraDependencies(ctx, queries, dbApp.ID)
	if err != nil {
		return nil, err
	}
	app := parseUserApplication(dbApp, deps)
	return &app, nil
}

func (svc *UserApplicationsService) GetAllUserApplications() ([]UserApplicationDao, error) {
	queries := infra_db_pg.New(svc.DbConn)
	ctx := context.Background()
	rows, err := queries.GetAllUserApplications(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all user applications: %w", err)
	}

	apps := make([]UserApplicationDao, 0, len(rows))
	for _, row := range rows {
		deps, err := getInfraDependencies(ctx, queries, row.ID)
		if err != nil {
			return nil, err
		}
		apps = append(apps, parseUserApplication(row, deps))
	}
	return apps, nil
}

func (svc *UserApplicationsService) UpdateUserApplication(id uuid.UUID, req UpdateUserApplicationRequest) (*UserApplicationDao, error) {
	queries := infra_db_pg.New(svc.DbConn)
	ctx := context.Background()
	current, err := queries.GetUserApplicationById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user application before update: %w", err)
	}

	deployConfig, err := marshalOptionalJSON(req.DeployConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal deploy config: %w", err)
	}
	buildConfig, err := marshalOptionalJSON(req.BuildConfig)
	if err != nil {
		return nil, fmt.Errorf("marshal build config: %w", err)
	}

	tx, err := svc.DbConn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	qtx := queries.WithTx(tx)

	name := current.Name
	if req.Name != "" {
		name = req.Name
	}
	repositoryURL := current.RepositoryUrl
	if req.RepositoryUrl != "" {
		repositoryURL = req.RepositoryUrl
	}
	sourceKind := current.SourceKind
	if req.SourceKind != "" {
		sourceKind = req.SourceKind
	}
	deployKind := current.DeployKind
	if req.DeployKind != "" {
		deployKind = req.DeployKind
	}
	registerable := current.Registerable
	if req.Registerable != nil {
		registerable = *req.Registerable
	}
	if deployConfig == nil {
		deployConfig = current.DeployConfig
	}
	if buildConfig == nil {
		buildConfig = current.BuildConfig
	}

	dbApp, err := qtx.UpdateUserApplication(ctx, infra_db_pg.UpdateUserApplicationParams{
		ID:             id,
		Name:           name,
		Description:    pgtype.Text{String: req.Description, Valid: req.Description != ""},
		RepositoryUrl:  repositoryURL,
		ManifestPath:   pgtype.Text{String: req.ManifestPath, Valid: req.ManifestPath != ""},
		SourceKind:     sourceKind,
		ModuleName:     pgtype.Text{String: req.ModuleName, Valid: req.ModuleName != ""},
		PackageName:    pgtype.Text{String: req.PackageName, Valid: req.PackageName != ""},
		PackageManager: pgtype.Text{String: req.PackageManager, Valid: req.PackageManager != ""},
		DeployKind:     deployKind,
		Registerable:   registerable,
		DeployConfig:   deployConfig,
		BuildConfig:    buildConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("update user application: %w", err)
	}

	if req.InfraDependencies != nil {
		if err = qtx.DeleteUserApplicationInfraDependenciesByAppId(ctx, id); err != nil {
			return nil, fmt.Errorf("delete user application dependencies: %w", err)
		}
	}
	deps := []InfraDependencyDao(nil)
	if req.InfraDependencies != nil {
		deps, err = createInfraDependencies(ctx, qtx, id, req.InfraDependencies)
		if err != nil {
			return nil, err
		}
	} else {
		deps, err = getInfraDependencies(ctx, qtx, id)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	app := parseUserApplication(dbApp, deps)
	return &app, nil
}

func (svc *UserApplicationsService) DeleteUserApplicationById(id uuid.UUID) error {
	return infra_db_pg.New(svc.DbConn).DeleteUserApplicationById(context.Background(), id)
}

func (svc *UserApplicationsService) DeleteUserApplicationByName(name string) error {
	return infra_db_pg.New(svc.DbConn).DeleteUserApplicationByName(context.Background(), name)
}

type dependencyQueryRunner interface {
	CreateUserApplicationInfraDependency(ctx context.Context, arg infra_db_pg.CreateUserApplicationInfraDependencyParams) (infra_db_pg.UserApplicationInfraDependency, error)
	GetUserApplicationInfraDependenciesByAppId(ctx context.Context, userApplicationID uuid.UUID) ([]infra_db_pg.GetUserApplicationInfraDependenciesByAppIdRow, error)
	DeleteUserApplicationInfraDependenciesByAppId(ctx context.Context, userApplicationID uuid.UUID) error
}

func createInfraDependencies(ctx context.Context, queries dependencyQueryRunner, appID uuid.UUID, deps []CreateInfraDependencyRequest) ([]InfraDependencyDao, error) {
	created := make([]InfraDependencyDao, 0, len(deps))
	for _, dep := range deps {
		cfg, err := marshalOptionalJSON(dep.Config)
		if err != nil {
			return nil, fmt.Errorf("marshal dependency config for %q: %w", dep.DependencyName, err)
		}
		if cfg == nil {
			cfg = []byte("{}")
		}
		row, err := queries.CreateUserApplicationInfraDependency(ctx, infra_db_pg.CreateUserApplicationInfraDependencyParams{
			UserApplicationID: appID,
			DependencyType:    dep.DependencyType,
			DependencyName:    dep.DependencyName,
			HostServerTypeID:  uuidToPgtype(dep.HostServerTypeId),
			PlatformTypeID:    uuidToPgtype(dep.PlatformTypeId),
			DependencyConfig:  cfg,
		})
		if err != nil {
			return nil, fmt.Errorf("create dependency %q: %w", dep.DependencyName, err)
		}
		createdDep := InfraDependencyDao{
			Id:             row.ID,
			DependencyType: row.DependencyType,
			DependencyName: row.DependencyName,
			Config:         unmarshalOptionalJSON(row.DependencyConfig),
			CreatedAt:      row.CreatedAt.Time,
			LastModified:   row.LastModified.Time,
		}
		if row.HostServerTypeID.Valid {
			id := uuid.UUID(row.HostServerTypeID.Bytes)
			createdDep.HostServerTypeId = &id
		}
		if row.PlatformTypeID.Valid {
			id := uuid.UUID(row.PlatformTypeID.Bytes)
			createdDep.PlatformTypeId = &id
		}
		created = append(created, createdDep)
	}
	return created, nil
}

func getInfraDependencies(ctx context.Context, queries dependencyQueryRunner, appID uuid.UUID) ([]InfraDependencyDao, error) {
	rows, err := queries.GetUserApplicationInfraDependenciesByAppId(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("get infra dependencies: %w", err)
	}
	return parseInfraDependencies(rows), nil
}

func boolValue(v *bool) bool {
	return v != nil && *v
}

func boolPtr(v bool) *bool {
	return &v
}

func uuidToPgtype(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}
