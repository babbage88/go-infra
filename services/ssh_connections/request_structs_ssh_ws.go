package ssh_connections

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocket message types
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// SSH Connection Request
// swagger:parameters createSshConnection
type SshConnectRequestWrapper struct {
	// in: body
	Body SshConnectionRequest `json:"body"`
}

// SSH Connection Request
// swagger:model SshConnectionRequestDetails
type SshConnectionRequest struct {
	// Host server ID to connect to
	// required: true
	// example: 123e4567-e89b-12d3-a456-426614174000
	HostServerID uuid.UUID `json:"hostServerId" validate:"required"`

	// Username to connect as on the remote server
	// required: true
	// example: admin
	Username string `json:"username" validate:"required"`
}

// SSH Connection Response
// swagger:model SshConnectionResponse
type SshConnectionResponse struct {
	// Unique connection identifier
	// example: 123e4567-e89b-12d3-a456-426614174000
	ConnectionID uuid.UUID `json:"connectionId"`

	// WebSocket URL for terminal communication
	// example: ws://localhost:8080/ssh/websocket/123e4567-e89b-12d3-a456-426614174000
	WebsocketURL string `json:"websocketUrl"`

	// Whether the connection was successful
	// example: true
	Success bool `json:"success"`

	// Error message if connection failed
	// example: SSH key not found
	Error string `json:"error,omitempty"`
}

// SSH Close Parameter
// swagger:parameters closeSshConnection
type SshCloseParam struct {
	// In: path
	CONNID uuid.UUID `json:"CONNID"`
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
