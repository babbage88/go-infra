package ssh_connections

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/babbage88/go-infra/api/authapi"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocket message types
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// SSH Connection Request
// swagger:model SshConnectionRequest
type SshConnectionRequest struct {
	// Host server ID to connect to
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174000
	HostServerID string `json:"hostServerId" validate:"required"`

	// Username to connect as on the remote server
	// required: true
	// example: admin
	Username string `json:"username" validate:"required"`
}

// SSH Connection Response
// swagger:model SshConnectionResponse
type SshConnectionResponse struct {
	// Unique connection identifier
	// example: abc123def456ghi789
	ConnectionID string `json:"connectionId"`

	// WebSocket URL for terminal communication
	// example: ws://localhost:8080/ssh/websocket/abc123def456ghi789
	WebsocketURL string `json:"websocketUrl"`

	// Whether the connection was successful
	// example: true
	Success bool `json:"success"`

	// Error message if connection failed
	// example: SSH key not found
	Error string `json:"error,omitempty"`
}

// SSH Connection Close Response
// swagger:model SshConnectionCloseResponse
type SshConnectionCloseResponse struct {
	// Success message
	// example: Connection closed successfully
	Message string `json:"message"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
	EnableCompression: true,
}

// HTTP handlers for standard library

// swagger:route POST /ssh/connect ssh createSshConnection
// Create a new SSH connection to a host server.
// responses:
//
//	200: SshConnectionResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	403: description:Access denied
//	404: description:Host server not found
//	500: description:Internal Server Error
func (m *SSHConnectionManager) CreateSSHConnectionHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req SshConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.HostServerID == "" {
		http.Error(w, "hostServerId is required", http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, err := authapi.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse host server ID
	hostServerID, err := uuid.Parse(req.HostServerID)
	if err != nil {
		http.Error(w, "Invalid host server ID format", http.StatusBadRequest)
		return
	}

	// Check if user has access to this host
	hasAccess, err := m.HasSSHAccessToHost(userID, hostServerID)
	if err != nil {
		slog.Error("Failed to check SSH access", "error", err)
		http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
		return
	}
	if !hasAccess {
		http.Error(w, "Access denied to this host", http.StatusForbidden)
		return
	}

	// Get host server info
	hostInfo, err := m.getHostServerInfo(hostServerID)
	if err != nil {
		http.Error(w, "Host server not found", http.StatusNotFound)
		return
	}

	// Get SSH key for this user/host combination
	sshKey, err := m.GetSSHKeyForHost(userID, hostServerID)
	if err != nil {
		slog.Error("Failed to get SSH key", "error", err)
		http.Error(w, "SSH key not found", http.StatusInternalServerError)
		return
	}

	// Generate connection ID
	connectionID := generateConnectionID()

	// Create session
	session := m.CreateSession(connectionID, userID, hostServerID, req.Username)

	// Connect to SSH server
	if err := session.Connect(hostInfo, sshKey, m.config); err != nil {
		m.RemoveSession(connectionID)
		slog.Error("Failed to connect to SSH server", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Track session in database
	clientIP := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	if err := m.trackSSHSession(connectionID, userID, hostServerID, req.Username, clientIP, userAgent); err != nil {
		slog.Error("Failed to track SSH session", "error", err)
	}

	// Return connection info
	scheme := "ws"
	if r.TLS != nil {
		scheme = "wss"
	}
	websocketURL := fmt.Sprintf("%s://%s/ssh/websocket/%s", scheme, r.Host, connectionID)

	response := SshConnectionResponse{
		ConnectionID: connectionID,
		WebsocketURL: websocketURL,
		Success:      true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// swagger:route DELETE /ssh/connect/{connectionId} ssh closeSshConnection
// Close an SSH connection.
// responses:
//
//	200: SshConnectionCloseResponse
//	400: description:Invalid connection ID
//	401: description:Unauthorized
//	403: description:Access denied
//	404: description:Session not found
//	500: description:Internal Server Error
func (m *SSHConnectionManager) CloseSSHConnectionHandler(w http.ResponseWriter, r *http.Request) {
	// Extract connection ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 || pathParts[len(pathParts)-2] != "connect" {
		http.Error(w, "Invalid connection ID", http.StatusBadRequest)
		return
	}
	connectionID := pathParts[len(pathParts)-1]

	if connectionID == "" {
		http.Error(w, "Connection ID is required", http.StatusBadRequest)
		return
	}

	session, exists := m.GetSession(connectionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Validate user permissions
	userID, err := authapi.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user owns this session
	if session.UserID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Mark session as inactive in database
	m.markSessionInactive(connectionID)

	// Close session
	session.Close()
	m.RemoveSession(connectionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SshConnectionCloseResponse{Message: "Connection closed successfully"})
}

// swagger:route GET /ssh/websocket/{connectionId} ssh sshWebSocket
// WebSocket endpoint for SSH terminal communication.
// responses:
//
//	101: description:Switching Protocols
//	400: description:Invalid connection ID
//	401: description:Unauthorized
//	404: description:Session not found
func (m *SSHConnectionManager) SSHWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Extract connection ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 || pathParts[len(pathParts)-2] != "websocket" {
		http.Error(w, "Invalid connection ID", http.StatusBadRequest)
		return
	}
	connectionID := pathParts[len(pathParts)-1]

	if connectionID == "" {
		http.Error(w, "Connection ID is required", http.StatusBadRequest)
		return
	}

	// Get session
	session, exists := m.GetSession(connectionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Validate user permissions
	userID, err := authapi.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user owns this session
	if session.UserID != userID {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Upgrade to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WebSocket upgrade failed", "error", err)
		return
	}

	// Set WebSocket connection
	session.SetWebSocket(ws)

	// Start data transfer
	session.StartDataTransfer()

	// Clean up when WebSocket closes
	defer func() {
		m.RemoveSession(connectionID)
		session.Close()
	}()
}

// Helper function to get client IP
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header first
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check for X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to remote address
	return r.RemoteAddr
}
