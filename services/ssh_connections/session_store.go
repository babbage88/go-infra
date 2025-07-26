package ssh_connections

import (
	"time"

	"github.com/google/uuid"
)

// SessionStore defines the interface for persistent SSH session management.
type SessionStore interface {
	CreateSession(session *SSHSession) error
	GetSession(id uuid.UUID) (*SSHSession, error)
	ListActiveSessions() ([]*SSHSession, error)
	RemoveSession(id uuid.UUID) error
	UpdateSessionActivity(id uuid.UUID, lastActivity time.Time) error
	MarkSessionInactive(id uuid.UUID) error
}
