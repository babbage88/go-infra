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
type SshConnectionRequest struct {
	HostServerID string `json:"hostServerId"`
	Username     string `json:"username"`
}

// SSH Connection Response
type SshConnectionResponse struct {
	ConnectionID string `json:"connectionId"`
	WebsocketURL string `json:"websocketUrl"`
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
	EnableCompression: true,
}

// HTTP handlers for standard library
func (m *SSHConnectionManager) CreateSSHConnectionHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req SshConnectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
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
		http.Error(w, "Invalid host server ID", http.StatusBadRequest)
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

func (m *SSHConnectionManager) CloseSSHConnectionHandler(w http.ResponseWriter, r *http.Request) {
	// Extract connection ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid connection ID", http.StatusBadRequest)
		return
	}
	connectionID := pathParts[len(pathParts)-1]

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
	json.NewEncoder(w).Encode(map[string]string{"message": "Connection closed successfully"})
}

// WebSocket handler for SSH communication
func (m *SSHConnectionManager) SSHWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Extract connection ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid connection ID", http.StatusBadRequest)
		return
	}
	connectionID := pathParts[len(pathParts)-1]

	// Get session
	session, exists := m.GetSession(connectionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
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
