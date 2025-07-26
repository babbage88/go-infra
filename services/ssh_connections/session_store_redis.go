package ssh_connections

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisSessionStore struct {
	rdb *redis.Client
}

func NewRedisSessionStore(rdb *redis.Client) *RedisSessionStore {
	return &RedisSessionStore{rdb: rdb}
}

func (s *RedisSessionStore) sessionKey(id uuid.UUID) string {
	return fmt.Sprintf("ssh_session:%s", id.String())
}

func (s *RedisSessionStore) CreateSession(session *SSHSession) error {
	ctx := context.Background()
	b, err := json.Marshal(session)
	if err != nil {
		return err
	}
	if err := s.rdb.Set(ctx, s.sessionKey(session.ID), b, 0).Err(); err != nil {
		return err
	}
	return s.rdb.SAdd(ctx, "ssh_sessions:active", session.ID.String()).Err()
}

func (s *RedisSessionStore) GetSession(id uuid.UUID) (*SSHSession, error) {
	ctx := context.Background()
	val, err := s.rdb.Get(ctx, s.sessionKey(id)).Result()
	if err != nil {
		return nil, err
	}
	var sess SSHSession
	if err := json.Unmarshal([]byte(val), &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *RedisSessionStore) ListActiveSessions() ([]*SSHSession, error) {
	ctx := context.Background()
	ids, err := s.rdb.SMembers(ctx, "ssh_sessions:active").Result()
	if err != nil {
		return nil, err
	}
	var sessions []*SSHSession
	for _, idStr := range ids {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		val, err := s.rdb.Get(ctx, s.sessionKey(id)).Result()
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

func (s *RedisSessionStore) RemoveSession(id uuid.UUID) error {
	ctx := context.Background()
	if err := s.rdb.Del(ctx, s.sessionKey(id)).Err(); err != nil {
		return err
	}
	return s.rdb.SRem(ctx, "ssh_sessions:active", id.String()).Err()
}

func (s *RedisSessionStore) UpdateSessionActivity(id uuid.UUID, lastActivity time.Time) error {
	ctx := context.Background()
	val, err := s.rdb.Get(ctx, s.sessionKey(id)).Result()
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
	return s.rdb.Set(ctx, s.sessionKey(id), b, 0).Err()
}

func (s *RedisSessionStore) MarkSessionInactive(id uuid.UUID) error {
	ctx := context.Background()
	return s.rdb.SRem(ctx, "ssh_sessions:active", id.String()).Err()
}
