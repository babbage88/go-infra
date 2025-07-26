package ssh_connections

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	valkey "github.com/valkey-io/valkey-go"
)

type ValkeySessionStore struct {
	client valkey.Client
}

func NewValkeySessionStore(client valkey.Client) *ValkeySessionStore {
	return &ValkeySessionStore{client: client}
}

func (s *ValkeySessionStore) sessionKey(id uuid.UUID) string {
	return fmt.Sprintf("ssh_session:%s", id.String())
}

func (s *ValkeySessionStore) CreateSession(session *SSHSession) error {
	ctx := context.Background()
	b, err := json.Marshal(session)
	if err != nil {
		return err
	}
	if err := s.client.Do(ctx, s.client.B().Set().Key(s.sessionKey(session.ID)).Value(valkey.BinaryString(b)).Build()).Error(); err != nil {
		return err
	}
	return s.client.Do(ctx, s.client.B().Sadd().Key("ssh_sessions:active").Member(session.ID.String()).Build()).Error()
}

func (s *ValkeySessionStore) GetSession(id uuid.UUID) (*SSHSession, error) {
	ctx := context.Background()
	val, err := s.client.Do(ctx, s.client.B().Get().Key(s.sessionKey(id)).Build()).ToString()
	if err != nil {
		return nil, err
	}
	var sess SSHSession
	if err := json.Unmarshal([]byte(val), &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *ValkeySessionStore) ListActiveSessions() ([]*SSHSession, error) {
	ctx := context.Background()
	ids, err := s.client.Do(ctx, s.client.B().Smembers().Key("ssh_sessions:active").Build()).AsStrSlice()
	if err != nil {
		return nil, err
	}
	var sessions []*SSHSession
	for _, idStr := range ids {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		val, err := s.client.Do(ctx, s.client.B().Get().Key(s.sessionKey(id)).Build()).ToString()
		if err != nil {
			continue
		}
		var sess SSHSession
		if err := json.Unmarshal([]byte(val), &sess); err != nil {
			continue
		}
		sessions = append(sessions, &sess)
	}
	return sessions, nil
}

func (s *ValkeySessionStore) RemoveSession(id uuid.UUID) error {
	ctx := context.Background()
	if err := s.client.Do(ctx, s.client.B().Del().Key(s.sessionKey(id)).Build()).Error(); err != nil {
		return err
	}
	return s.client.Do(ctx, s.client.B().Srem().Key("ssh_sessions:active").Member(id.String()).Build()).Error()
}

func (s *ValkeySessionStore) UpdateSessionActivity(id uuid.UUID, lastActivity time.Time) error {
	ctx := context.Background()
	val, err := s.client.Do(ctx, s.client.B().Get().Key(s.sessionKey(id)).Build()).ToString()
	if err != nil {
		return err
	}
	var sess SSHSession
	if err := json.Unmarshal([]byte(val), &sess); err != nil {
		return err
	}
	sess.LastActivity = lastActivity
	b, err := json.Marshal(&sess)
	if err != nil {
		return err
	}
	return s.client.Do(ctx, s.client.B().Set().Key(s.sessionKey(id)).Value(valkey.BinaryString(b)).Build()).Error()
}

func (s *ValkeySessionStore) MarkSessionInactive(id uuid.UUID) error {
	ctx := context.Background()
	return s.client.Do(ctx, s.client.B().Srem().Key("ssh_sessions:active").Member(id.String()).Build()).Error()
}
