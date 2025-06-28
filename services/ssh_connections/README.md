# SSH Connections Service

A complete Go backend implementation for creating WebSocket-based SSH connections to remote servers. This service allows browser terminals to securely connect to SSH servers through a WebSocket proxy.

## 🏗️ Architecture Overview

The SSH connection service creates a **WebSocket-based SSH proxy** that allows browser terminals to connect to remote servers securely:

```
┌─────────────┐    WebSocket    ┌─────────────┐    SSH    ┌─────────────┐
│   Browser   │ ←────────────→ │   Go Backend │ ←────────→ │ SSH Server  │
│ (xterm.js)  │                │             │            │             │
└─────────────┘                └─────────────┘            └─────────────┘
```

### **Detailed Flow:**
1. **User clicks "Connect"** → Frontend calls `POST /ssh/connect`
2. **Backend validates** → Checks JWT token, user permissions, SSH key access
3. **SSH connection created** → Backend establishes SSH connection to target server
4. **WebSocket URL returned** → Frontend receives `connectionId` and `websocketUrl`
5. **WebSocket connection** → Browser connects to WebSocket endpoint
6. **Bidirectional data flow** → Terminal I/O flows through WebSocket ↔ SSH
7. **Session cleanup** → Resources cleaned up on disconnect

## 📦 Dependencies

The service uses the following dependencies (already available in your `go.mod`):

- `github.com/gorilla/websocket` - WebSocket handling
- `golang.org/x/crypto/ssh` - SSH client implementation
- `github.com/google/uuid` - UUID generation
- `golang.org/x/time/rate` - Rate limiting
- Standard library `net/http` - HTTP server

## 🔐 Security Features

### **1. JWT Authentication**
- Uses existing `authapi.AuthMiddleware` for JWT validation
- Extracts user ID from JWT claims for authorization

### **2. Permission Checking**
- Validates user has SSH key access to specific host servers
- Checks SSH key host mappings in database
- Ensures users can only access authorized servers

### **3. Host Key Verification**
- Automatic known_hosts management
- Accepts and saves new host keys securely
- Prevents man-in-the-middle attacks

### **4. Rate Limiting**
- Per-user rate limiting to prevent abuse
- Configurable requests per second and burst limits

## 🗄️ Database Schema

The service uses the following database tables (already created via migrations):

### **ssh_sessions**
```sql
CREATE TABLE ssh_sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    host_server_id UUID NOT NULL REFERENCES host_servers(id) ON DELETE CASCADE,
    username VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    last_activity TIMESTAMP DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    client_ip INET,
    user_agent TEXT
);
```

### **ssh_connection_logs**
```sql
CREATE TABLE ssh_connection_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id),
    host_server_id UUID NOT NULL REFERENCES host_servers(id),
    action VARCHAR(50) NOT NULL, -- 'connect', 'disconnect', 'error'
    details JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## 🚀 Quick Start

### **1. Set up the SSH Connection Manager**

```go
package main

import (
    "context"
    "database/sql"
    "log/slog"
    "net/http"
    "os"
    "time"

    "github.com/babbage88/go-infra/api/authapi"
    "github.com/babbage88/go-infra/database/infra_db_pg"
    "github.com/babbage88/go-infra/services/ssh_connections"
    "github.com/babbage88/go-infra/services/user_secrets"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
    // 1. Set up database connections
    dbPool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
    if err != nil {
        slog.Error("Failed to connect to database", "error", err)
        return
    }
    defer dbPool.Close()

    // Create raw SQL connection for custom queries
    rawDB, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        slog.Error("Failed to create raw database connection", "error", err)
        return
    }
    defer rawDB.Close()

    // 2. Set up services
    queries := infra_db_pg.New(dbPool)
    secretProvider := user_secrets.NewPgUserSecretStore(dbPool)

    // 3. Configure SSH settings
    sshConfig := &ssh_connections.SSHConfig{
        KnownHostsPath: os.Getenv("SSH_KNOWN_HOSTS_PATH"),
        SSHTimeout:     30 * time.Second,
        MaxSessions:    100,
        RateLimit:      10, // requests per second
    }

    // 4. Create SSH connection manager
    sshManager := ssh_connections.NewSSHConnectionManager(queries, rawDB, secretProvider, sshConfig)

    // 5. Set up HTTP routes
    mux := http.NewServeMux()

    // SSH connection endpoints
    mux.Handle("POST /ssh/connect", authapi.AuthMiddleware(http.HandlerFunc(sshManager.CreateSSHConnectionHandler)))
    mux.Handle("DELETE /ssh/connect/{connectionId}", authapi.AuthMiddleware(http.HandlerFunc(sshManager.CloseSSHConnectionHandler)))
    mux.Handle("GET /ssh/websocket/{connectionId}", authapi.AuthMiddleware(http.HandlerFunc(sshManager.SSHWebSocketHandler)))

    // 6. Set up rate limiting (optional)
    rateLimiter := ssh_connections.NewRateLimiter(10, 5) // 10 requests per second, burst of 5
    rateLimitedMux := ssh_connections.RateLimitMiddleware(rateLimiter)(mux)

    // 7. Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    slog.Info("Starting SSH service", "port", port)
    if err := http.ListenAndServe(":"+port, rateLimitedMux); err != nil {
        slog.Error("Failed to start server", "error", err)
    }
}
```

### **2. Environment Variables**

```bash
# Database
DATABASE_URL=postgres://user:pass@localhost/dbname?sslmode=disable

# SSH Configuration
SSH_KNOWN_HOSTS_PATH=/root/.ssh/known_hosts
SSH_TIMEOUT=30s
MAX_SESSIONS=100
RATE_LIMIT=10

# Server
PORT=8080
```

## 📡 API Endpoints

### **1. Create SSH Connection**
```http
POST /ssh/connect
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "hostServerId": "123e4567-e89b-12d3-a456-426614174000",
  "username": "admin",
  "columns": 120,
  "rows": 30
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

### **2. WebSocket Connection**
```http
GET /ssh/websocket/{connectionId}
Authorization: Bearer <jwt_token>
```

**WebSocket Messages:**

**From Client:**
```json
{"type": "input", "data": "ls -la\n"}
{"type": "resize", "data": {"cols": 80, "rows": 24}}
```

**From Server:**
```json
{"type": "data", "data": "total 8\ndrwxr-xr-x 2 user user 4096 Jan 1 12:00 ."}
{"type": "error", "data": "command not found"}
```

### **3. Close SSH Connection**
```http
DELETE /ssh/connect/{connectionId}
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "message": "Connection closed successfully"
}
```

## 🔧 Configuration

### **SSHConfig**
```go
type SSHConfig struct {
    KnownHostsPath string        // Path to known_hosts file
    SSHTimeout     time.Duration // SSH connection timeout
    MaxSessions    int           // Maximum concurrent sessions
    RateLimit      int           // Requests per second for rate limiting
}
```

### **Rate Limiting**
```go
// Create rate limiter: 10 requests per second, burst of 5
rateLimiter := ssh_connections.NewRateLimiter(10, 5)

// Apply to HTTP handler
handler := ssh_connections.RateLimitMiddleware(rateLimiter)(mux)
```

## 🧪 Testing

### **Unit Tests**
```go
func TestCreateSSHConnection(t *testing.T) {
    // Setup mock database and SSH manager
    // Test connection creation
    // Verify response format
}
```

### **Integration Tests**
```go
func TestWebSocketCommunication(t *testing.T) {
    // Setup mock SSH server
    // Test WebSocket connection
    // Test data transfer
    // Test terminal resize
    // Test connection cleanup
}
```

## 🔍 Monitoring & Logging

### **Structured Logging**
The service uses `log/slog` for structured logging:

```go
slog.Info("SSH connection created", 
    "user_id", userID, 
    "host_server_id", hostServerID,
    "connection_id", connectionID)
```

### **Database Monitoring**
- Track active sessions in `ssh_sessions` table
- Monitor connection logs in `ssh_connection_logs` table
- Clean up expired sessions automatically

### **Health Checks**
```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "active_sessions": 5,
  "uptime": "2h30m15s"
}
```

## 🚨 Error Handling

### **Common Error Responses**

**400 Bad Request:**
```json
{"error": "Invalid host server ID"}
```

**401 Unauthorized:**
```json
{"error": "User not authenticated"}
```

**403 Forbidden:**
```json
{"error": "Access denied to this host"}
```

**404 Not Found:**
```json
{"error": "Session not found"}
```

**429 Too Many Requests:**
```json
{"error": "Rate limit exceeded"}
```

**500 Internal Server Error:**
```json
{"error": "SSH key not found"}
```

## 🔒 Security Best Practices

1. **Always use HTTPS/WSS in production**
2. **Implement proper origin checking for WebSocket connections**
3. **Regularly rotate SSH keys**
4. **Monitor connection logs for suspicious activity**
5. **Set appropriate rate limits**
6. **Use strong JWT secrets**
7. **Implement session timeouts**

## 📚 Integration with Frontend

### **TypeScript Client Example**
```typescript
import { SshConnectionService } from './generated-client';

// Create SSH connection
const response = await SshConnectionService.createSshConnection({
  hostServerId: "123e4567-e89b-12d3-a456-426614174000",
  username: "admin",
  columns: 120,
  rows: 30
});

// Connect to WebSocket
const ws = new WebSocket(response.websocketUrl);

// Send terminal input
ws.send(JSON.stringify({
  type: "input",
  data: "ls -la\n"
}));

// Handle terminal output
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  if (message.type === "data") {
    terminal.write(message.data);
  }
};

// Close connection
await SshConnectionService.closeSshConnection(response.connectionId);
```

## 🤝 Contributing

1. Follow the existing code style and patterns
2. Add tests for new functionality
3. Update documentation for API changes
4. Ensure security best practices are followed

## 📄 License

This service is part of the go-infra project and follows the same licensing terms. 

## Next Steps: Diagnosing the Interactive Shell

### 1. **Try a Minimal Shell**
Instead of s.SSHSession.Shell(), try running `/bin/bash -i` or `/bin/sh -i` as a command to see if an interactive shell will stay open.

### 2. **Check for Forced Commands or Restricted Shells**
- Ensure the SSH server is not configured to force a command or restrict the shell for this user/key.

### 3. **Check Remote Shell Configs**
- Temporarily move `.bashrc`, `.profile`, etc. out of the way for the SSH user and try again.

### 4. **Restore Interactive Shell Attempt**
- After the above, restore the code to use s.SSHSession.Shell() and see if the shell stays open.

## Code Suggestion: Try an Interactive Shell as a Command

Let's try running `/bin/bash -i` as a command instead of Shell():

```go
if err := s.SSHSession.Start("/bin/bash", "-i"); err != nil {
    slog.Error("Failed to start interactive bash", "error", err)
    return
}
```

Or for sh:
```go
if err := s.SSHSession.Start("/bin/sh", "-i"); err != nil {
    slog.Error("Failed to start interactive sh", "error", err)
    return
}
```

Would you like me to update your code to try this approach?  
Or do you want to try restoring Shell() and checking the remote shell configs first? 