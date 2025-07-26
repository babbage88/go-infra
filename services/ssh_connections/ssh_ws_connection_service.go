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
	store          SessionStore
	db             *infra_db_pg.Queries
	pool           *pgxpool.Pool
	config         *SSHConfig
	secretProvider user_secrets.UserSecretProvider

	// In-memory map for live sessions (per pod)
	liveSessions map[uuid.UUID]*SSHSession
	mu           sync.RWMutex
}

type SSHConfig struct {
	KnownHostsPath string
	SSHTimeout     time.Duration
	MaxSessions    int
	RateLimit      int // requests per second
}

func NewSSHConnectionManager(store SessionStore, db *infra_db_pg.Queries, pool *pgxpool.Pool, secretProvider user_secrets.UserSecretProvider, config *SSHConfig) *SSHConnectionManager {
	manager := &SSHConnectionManager{
		store:          store,
		db:             db,
		pool:           pool,
		config:         config,
		secretProvider: secretProvider,
		liveSessions:   make(map[uuid.UUID]*SSHSession),
	}

	// Start cleanup goroutine
	go manager.cleanupExpiredSessions()

	return manager
}

// Generate random connection ID
func generateConnectionID() uuid.UUID {
	return uuid.New()
}

// Create new SSH session (live + persistent)
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
	// Store in-memory
	m.mu.Lock()
	m.liveSessions[id] = session
	m.mu.Unlock()
	// Persist metadata
	_ = m.store.CreateSession(session)
	return session
}

// Get session: prefer live (in-memory), fallback to persistent (metadata only)
func (m *SSHConnectionManager) GetSession(id uuid.UUID) (*SSHSession, bool) {
	m.mu.RLock()
	session, ok := m.liveSessions[id]
	m.mu.RUnlock()
	if ok {
		return session, true
	}
	// Fallback: get from persistent store (metadata only, not live connection)
	meta, err := m.store.GetSession(id)
	if err != nil || meta == nil {
		return nil, false
	}
	return meta, false // not a live session
}

// Remove session from both in-memory and persistent store
func (m *SSHConnectionManager) RemoveSession(id uuid.UUID) {
	m.mu.Lock()
	delete(m.liveSessions, id)
	m.mu.Unlock()
	_ = m.store.RemoveSession(id)
}

// Cleanup expired sessions
func (m *SSHConnectionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Optionally, implement cleanup logic in a specific store implementation if needed
		// m.store.CleanupExpiredSessions() // <-- Remove this line

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
func (m *SSHConnectionManager) markSessionInactive(id uuid.UUID) {
	_ = m.store.MarkSessionInactive(id)
}

// List all active SSH sessions
func (m *SSHConnectionManager) ListActiveSessions() []*SSHSession {
	sessions, _ := m.store.ListActiveSessions()
	return sessions
}

func (m *SSHConnectionManager) updateSessionActivity(id uuid.UUID, lastActivity time.Time) {
	_ = m.store.UpdateSessionActivity(id, lastActivity)
}

// Rehydrate and reconnect a session from persistent store
func (m *SSHConnectionManager) RehydrateSessionAndConnect(meta *SSHSession, columns, rows int) (*SSHSession, error) {
	hostInfo, err := m.getHostServerInfo(meta.HostServerID)
	if err != nil {
		return nil, fmt.Errorf("host info not found: %w", err)
	}
	sshKey, err := m.GetSSHKeyForHost(meta.UserID, meta.HostServerID)
	if err != nil {
		return nil, fmt.Errorf("ssh key not found: %w", err)
	}
	// Create a new in-memory session
	session := m.CreateSession(meta.ID, meta.UserID, meta.HostServerID, meta.Username)
	if err := session.Connect(hostInfo, sshKey, m.config, columns, rows); err != nil {
		m.RemoveSession(meta.ID)
		return nil, fmt.Errorf("failed to reconnect SSH: %w", err)
	}
	return session, nil
}
