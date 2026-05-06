package user_applications

import (
	"context"
	"testing"
	"time"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type stubDependencyQueryRunner struct {
	createCalls []infra_db_pg.CreateUserApplicationInfraDependencyParams
}

func (s *stubDependencyQueryRunner) CreateUserApplicationInfraDependency(_ context.Context, arg infra_db_pg.CreateUserApplicationInfraDependencyParams) (infra_db_pg.UserApplicationInfraDependency, error) {
	createdAt := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	s.createCalls = append(s.createCalls, arg)
	return infra_db_pg.UserApplicationInfraDependency{
		ID:               uuid.New(),
		UserApplicationID: arg.UserApplicationID,
		DependencyType:   arg.DependencyType,
		DependencyName:   arg.DependencyName,
		HostServerTypeID: arg.HostServerTypeID,
		PlatformTypeID:   arg.PlatformTypeID,
		DependencyConfig: arg.DependencyConfig,
		CreatedAt:        createdAt,
		LastModified:     createdAt,
	}, nil
}

func (s *stubDependencyQueryRunner) GetUserApplicationInfraDependenciesByAppId(context.Context, uuid.UUID) ([]infra_db_pg.GetUserApplicationInfraDependenciesByAppIdRow, error) {
	return nil, nil
}

func (s *stubDependencyQueryRunner) DeleteUserApplicationInfraDependenciesByAppId(context.Context, uuid.UUID) error {
	return nil
}

func TestCreateInfraDependenciesDefaultsNilConfigToEmptyJSON(t *testing.T) {
	t.Parallel()

	runner := &stubDependencyQueryRunner{}
	appID := uuid.New()

	created, err := createInfraDependencies(context.Background(), runner, appID, []CreateInfraDependencyRequest{
		{
			DependencyType: "host_server_type",
			DependencyName: "Application Server",
		},
	})
	if err != nil {
		t.Fatalf("createInfraDependencies returned error: %v", err)
	}

	if len(runner.createCalls) != 1 {
		t.Fatalf("expected 1 create call, got %d", len(runner.createCalls))
	}
	if got := string(runner.createCalls[0].DependencyConfig); got != "{}" {
		t.Fatalf("expected dependency config to default to {}, got %q", got)
	}
	if len(created) != 1 {
		t.Fatalf("expected 1 created dependency, got %d", len(created))
	}
	if created[0].Config == nil {
		t.Fatalf("expected created dependency config to be non-nil")
	}
}
