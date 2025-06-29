package ssh_connections

import (
	"context"
	"time"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type DBSessionStore struct {
	db *infra_db_pg.Queries
}

func NewDBSessionStore(db *infra_db_pg.Queries) *DBSessionStore {
	return &DBSessionStore{db: db}
}

func toPgTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func fromPgTime(p pgtype.Timestamptz) time.Time {
	if p.Valid {
		return p.Time
	}
	return time.Time{}
}

func (s *DBSessionStore) CreateSession(session *SSHSession) error {
	return s.db.CreateSSHSession(context.Background(), infra_db_pg.CreateSSHSessionParams{
		ID:           session.ID,
		UserID:       session.UserID,
		HostServerID: session.HostServerID,
		Username:     session.Username,
		CreatedAt:    toPgTime(session.CreatedAt),
		LastActivity: toPgTime(session.LastActivity),
	})
}

func (s *DBSessionStore) GetSession(id uuid.UUID) (*SSHSession, error) {
	dbSess, err := s.db.GetSSHSessionById(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return &SSHSession{
		ID:           dbSess.ID,
		UserID:       dbSess.UserID,
		HostServerID: dbSess.HostServerID,
		Username:     dbSess.Username,
		CreatedAt:    fromPgTime(dbSess.CreatedAt),
		LastActivity: fromPgTime(dbSess.LastActivity),
	}, nil
}

func (s *DBSessionStore) ListActiveSessions() ([]*SSHSession, error) {
	dbSessions, err := s.db.ListActiveSSHSessions(context.Background())
	if err != nil {
		return nil, err
	}
	sessions := make([]*SSHSession, 0, len(dbSessions))
	for _, dbSess := range dbSessions {
		sessions = append(sessions, &SSHSession{
			ID:           dbSess.ID,
			UserID:       dbSess.UserID,
			HostServerID: dbSess.HostServerID,
			Username:     dbSess.Username,
			CreatedAt:    fromPgTime(dbSess.CreatedAt),
			LastActivity: fromPgTime(dbSess.LastActivity),
		})
	}
	return sessions, nil
}

func (s *DBSessionStore) RemoveSession(id uuid.UUID) error {
	return s.db.RemoveSSHSession(context.Background(), id)
}

func (s *DBSessionStore) UpdateSessionActivity(id uuid.UUID, lastActivity time.Time) error {
	return s.db.UpdateSSHSessionActivity(context.Background(), infra_db_pg.UpdateSSHSessionActivityParams{
		ID:           id,
		LastActivity: toPgTime(lastActivity),
	})
}

func (s *DBSessionStore) MarkSessionInactive(id uuid.UUID) error {
	return s.db.MarkSSHSessionInactive(context.Background(), id)
}
