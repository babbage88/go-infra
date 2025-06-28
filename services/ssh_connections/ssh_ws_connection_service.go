package ssh_connections

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/ssh"
)

// SSH Session represents an active SSH connection
type SSHSession struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	HostServerID uuid.UUID
	Username     string
	SSHClient    *ssh.Client
	SSHSession   *ssh.Session
	WebSocket    *websocket.Conn
	CreatedAt    time.Time
	LastActivity time.Time
	mu           sync.Mutex
	db           *infra_db_pg.Queries
	dbtx         infra_db_pg.DBTX
}

type SSHConnectionLog struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	SessionID    string          `json:"session_id" db:"session_id"`
	UserID       uuid.UUID       `json:"user_id" db:"user_id"`
	HostServerID uuid.UUID       `json:"host_server_id" db:"host_server_id"`
	Action       string          `json:"action" db:"action"`
	Details      json.RawMessage `json:"details" db:"details"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

type SSHKeyInfo struct {
	ID         uuid.UUID `json:"id"`
	PrivateKey string    `json:"private_key"`
	PublicKey  string    `json:"public_key"`
	Passphrase string    `json:"passphrase"`
	KeyType    string    `json:"key_type"`
	Username   string    `json:"username"`
}

type HostServerInfo struct {
	ID        uuid.UUID `json:"id"`
	Hostname  string    `json:"hostname"`
	IPAddress string    `json:"ip_address"`
	Username  string    `json:"username"`
	Port      int       `json:"port"`
}

// SSH Connection Manager
type SSHConnectionManager struct {
	sessions       map[uuid.UUID]*SSHSession
	mu             sync.RWMutex
	db             *infra_db_pg.Queries
	pool           *pgxpool.Pool
	config         *SSHConfig
	secretProvider user_secrets.UserSecretProvider
}

type SSHConfig struct {
	KnownHostsPath string
	SSHTimeout     time.Duration
	MaxSessions    int
	RateLimit      int // requests per second
}

func NewSSHConnectionManager(db *infra_db_pg.Queries, pool *pgxpool.Pool, secretProvider user_secrets.UserSecretProvider, config *SSHConfig) *SSHConnectionManager {
	manager := &SSHConnectionManager{
		sessions:       make(map[uuid.UUID]*SSHSession),
		db:             db,
		pool:           pool,
		config:         config,
		secretProvider: secretProvider,
	}

	// Start cleanup goroutine
	go manager.cleanupExpiredSessions()

	return manager
}

// Generate random connection ID
func generateConnectionID() uuid.UUID {
	return uuid.New()
}

// Create new SSH session
func (m *SSHConnectionManager) CreateSession(id uuid.UUID, userID uuid.UUID, hostServerID uuid.UUID, username string) *SSHSession {
	session := &SSHSession{
		ID:           id,
		UserID:       userID,
		HostServerID: hostServerID,
		Username:     username,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		db:           m.db,
		dbtx:         m.pool,
	}

	m.mu.Lock()
	m.sessions[id] = session
	m.mu.Unlock()

	return session
}

// Get session by ID
func (m *SSHConnectionManager) GetSession(id uuid.UUID) (*SSHSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	session, exists := m.sessions[id]
	return session, exists
}

// Remove session
func (m *SSHConnectionManager) RemoveSession(id uuid.UUID) {
	m.mu.Lock()
	if session, exists := m.sessions[id]; exists {
		session.Close()
		delete(m.sessions, id)
	}
	m.mu.Unlock()
}

// Cleanup expired sessions
func (m *SSHConnectionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for id, session := range m.sessions {
			// Close sessions older than 1 hour
			if now.Sub(session.CreatedAt) > time.Hour {
				session.Close()
				delete(m.sessions, id)
			}
		}
		m.mu.Unlock()

		// Clean up database sessions
		m.cleanupDatabaseSessions()
	}
}

func (m *SSHConnectionManager) cleanupDatabaseSessions() {
	ctx := context.Background()
	_, err := m.pool.Exec(ctx, `
        UPDATE ssh_sessions 
        SET is_active = false 
        WHERE last_activity < NOW() - INTERVAL '1 hour'
    `)
	if err != nil {
		// Log error if needed
	}
}

// Check if user has SSH access to specific host
func (m *SSHConnectionManager) HasSSHAccessToHost(userID, hostServerID uuid.UUID) (bool, error) {
	mappings, err := m.db.GetSSHKeyHostMappingsByHostId(context.Background(), hostServerID)
	if err != nil {
		return false, fmt.Errorf("failed to check SSH access: %w", err)
	}

	for _, mapping := range mappings {
		if mapping.UserID == userID {
			return true, nil
		}
	}

	return false, nil
}

// Get user's SSH key for specific host
func (m *SSHConnectionManager) GetSSHKeyForHost(userID, hostServerID uuid.UUID) (*SSHKeyInfo, error) {
	mappings, err := m.db.GetSSHKeyHostMappingsByHostId(context.Background(), hostServerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH key mappings: %w", err)
	}

	for _, mapping := range mappings {
		if mapping.UserID == userID {
			// Get SSH key details
			sshKey, err := m.db.GetSSHKeyById(context.Background(), mapping.SshKeyID)
			if err != nil {
				return nil, fmt.Errorf("failed to get SSH key: %w", err)
			}

			// Get private key from secrets
			var privateKey string
			if sshKey.PrivSecretID != uuid.Nil {
				secret, err := m.secretProvider.RetrieveSecret(sshKey.PrivSecretID)
				if err != nil {
					return nil, fmt.Errorf("failed to retrieve SSH key secret: %w", err)
				}
				privateKey = string(secret.ExternalAuthToken.Token)
			}

			var passphrase string
			// Get passphrase if key needs one from secrets
			if sshKey.PassphraseID != nil {
				secret, err := m.secretProvider.RetrieveSecret(*sshKey.PassphraseID)
				if err != nil {
					return nil, fmt.Errorf("failed to retrieve SSH key passphrase: %w", err)
				}
				passphrase = string(secret.ExternalAuthToken.Token)
			}

			return &SSHKeyInfo{
				ID:         sshKey.ID,
				PrivateKey: privateKey,
				Passphrase: passphrase,
				PublicKey:  sshKey.PublicKey,
				KeyType:    sshKey.KeyType,
				Username:   mapping.HostserverUsername,
			}, nil
		}
	}

	return nil, fmt.Errorf("SSH key not found for user and host")
}

// Get host server info
func (m *SSHConnectionManager) getHostServerInfo(hostServerID uuid.UUID) (*HostServerInfo, error) {
	server, err := m.db.GetHostServerById(context.Background(), hostServerID)
	if err != nil {
		return nil, fmt.Errorf("host server not found: %w", err)
	}

	return &HostServerInfo{
		ID:        server.ID,
		Hostname:  server.Hostname,
		IPAddress: server.IpAddress.String(),
		Username:  "", // Will be set from SSH key mapping
		Port:      22, // Default SSH port
	}, nil
}

// Track SSH session in database
func (m *SSHConnectionManager) trackSSHSession(sessionID uuid.UUID, userID, hostServerID uuid.UUID, username, clientIP, userAgent string) error {
	query := `
        INSERT INTO ssh_sessions (id, user_id, host_server_id, username, client_ip, user_agent, is_active, created_at, last_activity)
        VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())
        ON CONFLICT (id) DO UPDATE SET
            last_activity = NOW(),
            is_active = true,
            client_ip = $5,
            user_agent = $6
    `
	_, err := m.pool.Exec(context.Background(), query, sessionID, userID, hostServerID, username, clientIP, userAgent)
	return err
}

// Mark session as inactive
func (m *SSHConnectionManager) markSessionInactive(sessionID uuid.UUID) error {
	query := `UPDATE ssh_sessions SET is_active = false WHERE id = $1`
	_, err := m.pool.Exec(context.Background(), query, sessionID)
	return err
}
