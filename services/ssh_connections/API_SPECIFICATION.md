# SSH Connections API Specification

This document outlines the SSH Connections API endpoints and their implementation in the `ssh_connections` package.

## API Endpoints

### 1. Create SSH Connection

**Endpoint:** `POST /ssh/connect`

**Request Body:**
```json
{
  "hostServerId": "123e4567-e89b-12d3-a456-426614174000",
  "username": "admin"
}
```

**Response:**
```json
{
  "connectionId": "123e4567-e89b-12d3-a456-426614174000",
  "websocketUrl": "ws://localhost:8080/ssh/websocket/123e4567-e89b-12d3-a456-426614174000",
  "success": true
}
```

**Implementation:** `CreateSSHConnectionHandler` in `ssh_ws.go`

**Features:**
- Validates request body with required fields
- Checks user authentication and permissions
- Verifies SSH access to the specified host server
- Creates SSH session and establishes connection
- Tracks session in database with client IP and user agent
- Returns connection ID and WebSocket URL

### 2. Close SSH Connection

**Endpoint:** `DELETE /ssh/connect/{connectionId}`

**Path Parameters:**
- `connectionId` (UUID, required): The connection ID to close

**Response:**
```json
{
  "message": "Connection closed successfully"
}
```

**Implementation:** `CloseSSHConnectionHandler` in `ssh_ws.go`

**Features:**
- Extracts connection ID from URL path
- Validates user ownership of the session
- Marks session as inactive in database
- Closes SSH connection and cleans up resources

### 3. SSH WebSocket Connection

**Endpoint:** `GET /ssh/websocket/{connectionId}`

**Path Parameters:**
- `connectionId` (UUID, required): The connection ID for WebSocket upgrade

**Response:** Upgrades to WebSocket connection for terminal communication

**Implementation:** `SSHWebSocketHandler` in `ssh_ws.go`

**Features:**
- Extracts connection ID from URL path
- Validates user ownership of the session
- Upgrades HTTP connection to WebSocket
- Establishes bidirectional data transfer between client and SSH server
- Handles connection cleanup on close

## Data Models

### SshConnectionRequest
```go
type SshConnectionRequest struct {
    HostServerID uuid.UUID `json:"hostServerId" validate:"required"`
    Username     string    `json:"username" validate:"required"`
}
```

### SshConnectionResponse
```go
type SshConnectionResponse struct {
    ConnectionID  uuid.UUID `json:"connectionId"`
    WebsocketURL  string    `json:"websocketUrl"`
    Success       bool      `json:"success"`
    Error         string    `json:"error,omitempty"`
}
```

### SshConnectionCloseResponse
```go
type SshConnectionCloseResponse struct {
    Message string `json:"message"`
}
```

## Authentication & Authorization

All endpoints require authentication via JWT token in the `Authorization` header:
```
Authorization: Bearer <jwt_token>
```

The implementation uses `authapi.AuthMiddleware` to:
- Extract and validate JWT tokens
- Store user claims in request context
- Ensure only authenticated users can access SSH connections

## Permission Checks

The implementation includes several permission checks:
1. **SSH Access Verification:** Users must have SSH key mappings to the target host server
2. **Session Ownership:** Users can only access their own SSH sessions
3. **Host Server Access:** Validates that the host server exists and is accessible

## Database Integration

The SSH connections are tracked in the database using:
- `ssh_sessions` table for active session tracking
- `ssh_connection_logs` table for audit logging
- Session status updates (active/inactive)
- Client IP and user agent tracking

## Error Handling

The API provides comprehensive error handling:
- **400 Bad Request:** Invalid request body, missing parameters, invalid UUID format
- **401 Unauthorized:** Missing or invalid JWT token
- **403 Forbidden:** Access denied to host server or session
- **404 Not Found:** Host server or session not found
- **500 Internal Server Error:** Database errors, SSH connection failures

## Usage Example

```go
// Set up SSH connection manager
config := &SSHConfig{
    KnownHostsPath: "/etc/ssh/known_hosts",
    SSHTimeout:     30,
    MaxSessions:    100,
    RateLimit:      10,
}

sshManager := NewSSHConnectionManager(db, pool, secretProvider, config)

// Set up routes with authentication
http.Handle("/ssh/connect", authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    sshManager.CreateSSHConnectionHandler(w, r)
})))

http.Handle("/ssh/connect/", authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodDelete {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    sshManager.CloseSSHConnectionHandler(w, r)
})))

http.Handle("/ssh/websocket/", authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    sshManager.SSHWebSocketHandler(w, r)
})))
```

## Security Features

1. **JWT Authentication:** All requests require valid JWT tokens
2. **Session Isolation:** Users can only access their own SSH sessions
3. **Permission Validation:** SSH access is verified against database mappings
4. **Input Validation:** All request parameters are validated
5. **Rate Limiting:** Configurable rate limiting per user
6. **Audit Logging:** All SSH connections are logged for security monitoring
7. **Connection Cleanup:** Automatic cleanup of expired sessions

## WebSocket Communication

The WebSocket endpoint provides:
- Real-time bidirectional communication
- SSH terminal data transfer
- Automatic connection cleanup
- Error handling and recovery
- Compression support for better performance 