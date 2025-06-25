package ssh_connections

import (
	"net/http"

	"github.com/babbage88/go-infra/api/authapi"
	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ExampleSetup shows how to set up the SSH connection manager with HTTP routes
func ExampleSetup(db *infra_db_pg.Queries, pool *pgxpool.Pool, secretProvider user_secrets.UserSecretProvider) {
	// Create SSH connection manager
	config := &SSHConfig{
		KnownHostsPath: "/etc/ssh/known_hosts",
		SSHTimeout:     30, // seconds
		MaxSessions:    100,
		RateLimit:      10, // requests per second
	}

	sshManager := NewSSHConnectionManager(db, pool, secretProvider, config)

	// Set up HTTP routes
	http.HandleFunc("/ssh/connect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		sshManager.CreateSSHConnectionHandler(w, r)
	})

	http.HandleFunc("/ssh/connect/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		sshManager.CloseSSHConnectionHandler(w, r)
	})

	http.HandleFunc("/ssh/websocket/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		sshManager.SSHWebSocketHandler(w, r)
	})
}

// ExampleWithMiddleware shows how to set up routes with authentication middleware
func ExampleWithMiddleware(db *infra_db_pg.Queries, pool *pgxpool.Pool, secretProvider user_secrets.UserSecretProvider) {
	// Create SSH connection manager
	config := &SSHConfig{
		KnownHostsPath: "/etc/ssh/known_hosts",
		SSHTimeout:     30,
		MaxSessions:    100,
		RateLimit:      10,
	}

	sshManager := NewSSHConnectionManager(db, pool, secretProvider, config)

	// Set up routes with authentication middleware
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
}

/*
API Endpoints Summary:

POST /ssh/connect
- Request Body: {"hostServerId": "uuid", "username": "string"}
- Response: {"connectionId": "string", "websocketUrl": "string", "success": true}

DELETE /ssh/connect/{connectionId}
- Path Parameter: connectionId (string)
- Response: {"message": "Connection closed successfully"}

GET /ssh/websocket/{connectionId}
- Path Parameter: connectionId (string)
- Upgrades to WebSocket connection for terminal communication

All endpoints require authentication via the authapi.AuthMiddleware.
*/
