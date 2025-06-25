package ssh_connections

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/babbage88/go-infra/api/authapi"
	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ExampleSetup shows how to set up and use the SSH connections package
func ExampleSetup() {
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
	sshConfig := &SSHConfig{
		KnownHostsPath: os.Getenv("SSH_KNOWN_HOSTS_PATH"),
		SSHTimeout:     30 * time.Second,
		MaxSessions:    100,
		RateLimit:      10, // requests per second
	}

	// 4. Create SSH connection manager
	sshManager := NewSSHConnectionManager(queries, rawDB, secretProvider, sshConfig)

	// 5. Set up HTTP routes
	mux := http.NewServeMux()

	// SSH connection endpoints - convert handlers to http.Handler
	mux.Handle("POST /ssh/connect", authapi.AuthMiddleware(http.HandlerFunc(sshManager.CreateSSHConnectionHandler)))
	mux.Handle("DELETE /ssh/connect/{connectionId}", authapi.AuthMiddleware(http.HandlerFunc(sshManager.CloseSSHConnectionHandler)))
	mux.Handle("GET /ssh/websocket/{connectionId}", authapi.AuthMiddleware(http.HandlerFunc(sshManager.SSHWebSocketHandler)))

	// 6. Set up rate limiting (optional)
	rateLimiter := NewRateLimiter(10, 5) // 10 requests per second, burst of 5
	rateLimitedMux := RateLimitMiddleware(rateLimiter)(mux)

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

// ExampleUsage shows how to use the SSH connections from a client perspective
func ExampleUsage() {
	// This is how a client would use the SSH connection service

	// 1. Create SSH connection request
	// request := SshConnectionRequest{
	//     HostServerID: "123e4567-e89b-12d3-a456-426614174000", // UUID of host server
	//     Username:     "admin",                                 // Username to connect as
	// }

	// 2. Send POST request to /ssh/connect with JWT token
	// The response will contain:
	// {
	//   "connectionId": "abc123...",
	//   "websocketUrl": "ws://localhost:8080/ssh/websocket/abc123...",
	//   "success": true
	// }

	// 3. Connect to WebSocket URL
	// The WebSocket will handle bidirectional SSH communication

	// 4. Send WebSocket messages for terminal interaction:
	// - {"type": "input", "data": "ls -la\n"} - Send command input
	// - {"type": "resize", "data": {"cols": 80, "rows": 24}} - Resize terminal

	// 5. Receive WebSocket messages:
	// - {"type": "data", "data": "output from command"} - Terminal output
	// - {"type": "error", "data": "error message"} - Error output

	// 6. Close connection by sending DELETE request to /ssh/connect/{connectionId}
}

// ExampleWithCustomMiddleware shows how to add custom middleware
func ExampleWithCustomMiddleware() {
	// Set up database and services (same as above)
	dbPool, _ := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	defer dbPool.Close()

	rawDB, _ := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	defer rawDB.Close()

	queries := infra_db_pg.New(dbPool)
	secretProvider := user_secrets.NewPgUserSecretStore(dbPool)

	sshConfig := &SSHConfig{
		KnownHostsPath: os.Getenv("SSH_KNOWN_HOSTS_PATH"),
		SSHTimeout:     30 * time.Second,
		MaxSessions:    100,
		RateLimit:      10,
	}

	sshManager := NewSSHConnectionManager(queries, rawDB, secretProvider, sshConfig)

	// Create custom middleware chain
	mux := http.NewServeMux()

	// Add SSH routes
	mux.Handle("POST /ssh/connect", http.HandlerFunc(sshManager.CreateSSHConnectionHandler))
	mux.Handle("DELETE /ssh/connect/{connectionId}", http.HandlerFunc(sshManager.CloseSSHConnectionHandler))
	mux.Handle("GET /ssh/websocket/{connectionId}", http.HandlerFunc(sshManager.SSHWebSocketHandler))

	// Add health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	// Apply middleware chain
	handler := authapi.AuthMiddleware(
		RateLimitMiddleware(NewRateLimiter(10, 5))(
			loggingMiddleware(mux),
		),
	)

	// Start server
	slog.Info("Starting SSH service with custom middleware")
	http.ListenAndServe(":8080", handler)
}

// loggingMiddleware is an example of custom middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		slog.Info("Request started", "method", r.Method, "path", r.URL.Path)

		next.ServeHTTP(w, r)

		slog.Info("Request completed", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}
