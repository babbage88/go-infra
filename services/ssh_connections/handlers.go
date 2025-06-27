package ssh_connections

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/babbage88/go-infra/api/authapi"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

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
	if req.HostServerID == uuid.Nil {
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

	// Check if user has access to this host
	hasAccess, err := m.HasSSHAccessToHost(userID, req.HostServerID)
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
	hostInfo, err := m.getHostServerInfo(req.HostServerID)
	if err != nil {
		http.Error(w, "Host server not found", http.StatusNotFound)
		return
	}

	// Get SSH key for this user/host combination
	sshKey, err := m.GetSSHKeyForHost(userID, req.HostServerID)
	if err != nil {
		slog.Error("Failed to get SSH key", "error", err)
		http.Error(w, "SSH key not found", http.StatusInternalServerError)
		return
	}

	// Generate connection ID
	connectionID := generateConnectionID()

	// Create session
	session := m.CreateSession(connectionID, userID, req.HostServerID, req.Username)

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
	if err := m.trackSSHSession(connectionID, userID, req.HostServerID, req.Username, clientIP, userAgent); err != nil {
		slog.Error("Failed to track SSH session", "error", err)
	}

	// Return connection info
	scheme := "ws"
	if r.TLS != nil {
		scheme = "wss"
	}
	wsHostPort := os.Getenv("WEBSOCKET_HOSTPORT")
	if wsHostPort == "" {
		wsHostPort = "localhost:8090"
	}
	websocketURL := fmt.Sprintf("%s://%s/ssh/websocket/%s", scheme, wsHostPort, connectionID)

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
	connectionIDStr := r.PathValue("CONNID")

	if connectionIDStr == "" {
		http.Error(w, "Connection ID is required", http.StatusBadRequest)
		return
	}

	// Parse connection ID as UUID
	connectionID, err := uuid.Parse(connectionIDStr)
	if err != nil {
		http.Error(w, "Invalid connection ID format", http.StatusBadRequest)
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
	// Extract JWT from header or query param
	token := r.Header.Get("Authorization")
	if token == "" {
		token = r.URL.Query().Get("token")
		if token != "" {
			token = "Bearer " + token
		}
	}
	if token == "" {
		http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
		return
	}

	// Inline JWT validation logic
	if !strings.HasPrefix(token, "Bearer ") {
		http.Error(w, "Unauthorized: malformed Authorization header", http.StatusUnauthorized)
		return
	}
	jwtToken := strings.TrimPrefix(token, "Bearer ")
	secret := os.Getenv("JWT_KEY")
	if secret == "" {
		http.Error(w, "Unauthorized: JWT_KEY not set", http.StatusUnauthorized)
		return
	}
	tok, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		slog.Error("JWT parse error", "error", err)
		http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
		return
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Unauthorized: invalid token claims", http.StatusUnauthorized)
		return
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		http.Error(w, "Unauthorized: missing sub claim", http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(sub)
	if err != nil {
		http.Error(w, "Unauthorized: invalid user id", http.StatusUnauthorized)
		return
	}

	// Extract connection ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 || pathParts[len(pathParts)-2] != "websocket" {
		http.Error(w, "Invalid connection ID", http.StatusBadRequest)
		return
	}
	connectionIDStr := pathParts[len(pathParts)-1]

	if connectionIDStr == "" {
		http.Error(w, "Connection ID is required", http.StatusBadRequest)
		return
	}

	// Parse connection ID as UUID
	connectionID, err := uuid.Parse(connectionIDStr)
	if err != nil {
		http.Error(w, "Invalid connection ID format", http.StatusBadRequest)
		return
	}

	// Get session
	session, exists := m.GetSession(connectionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Validate user permissions
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

	// Block until the WebSocket is closed
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break // WebSocket closed or error occurred
		}
	}

	// Clean up after WebSocket closes
	m.RemoveSession(connectionID)
	session.Close()

	// Put claims in context
	ctx := context.WithValue(r.Context(), authapi.ClaimsContextKey, claims)
	r = r.WithContext(ctx)
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
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // fallback, but this may still cause DB error
	}
	return host
}
