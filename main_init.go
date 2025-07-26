package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	infra_db "github.com/babbage88/go-infra/database/infra_db"
	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/services/ssh_connections"
	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valkey-io/valkey-go"
)

func initPgConnPool() *pgxpool.Pool {
	connPool := infra_db.PgPoolInit()
	return connPool
}

func initializeSshConnMgr(connPool *pgxpool.Pool, secretProvider user_secrets.UserSecretProvider, timeoutSec int, maxSessions int, rateLimit int) *ssh_connections.SSHConnectionManager {

	// Initialize SSH session store based on environment variable
	var sessionStore ssh_connections.SessionStore
	storeType := os.Getenv("SSH_SESSION_STORE_TYPE")
	dbQueries := infra_db_pg.New(connPool)
	if storeType == "valkey" {
		// Use Valkey-backed session store
		valkeyAddr := os.Getenv("VALKEY_ADDR")
		if valkeyAddr == "" {
			valkeyAddr = "127.0.0.1:6379"
		}
		slog.Info("Using Valkey session store", slog.String("address", valkeyAddr))
		valkeyClient, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{valkeyAddr},
			AuthCredentialsFn: func(acc valkey.AuthCredentialsContext) (valkey.AuthCredentials, error) {
				return valkey.AuthCredentials{
					Username: os.Getenv("VALKEY_USER"),
					Password: os.Getenv("VALKEY_PASSWORD"),
				}, nil
			}})
		if err != nil {
			slog.Error("Failed to initialize Valkey client", slog.String("error", err.Error()))
			os.Exit(1)
		}
		// Test connection with PING command using the proper Valkey client pattern
		if err := valkeyClient.Do(context.Background(), valkeyClient.B().Ping().Build()).Error(); err != nil {
			slog.Error("Failed to ping Valkey server", slog.String("error", err.Error()))
			os.Exit(1)
		}
		slog.Info("Successfully pinged Valkey server")
		sessionStore = ssh_connections.NewValkeySessionStore(valkeyClient)
		slog.Info("Using Valkey session store", slog.String("address", valkeyAddr))
	} else {
		// Default to Postgres-backed session store
		sessionStore = ssh_connections.NewDBSessionStore(dbQueries)
		slog.Info("Using Postgres session store")
	}
	wsKnownHostsPath := os.Getenv("WS_KNOWN_HOSTS")
	if wsKnownHostsPath == "" {
		wsKnownHostsPath = "ws_known_hosts"
	}

	sshConnectionManager := ssh_connections.NewSSHConnectionManager(
		sessionStore,
		dbQueries,
		connPool,
		secretProvider,
		&ssh_connections.SSHConfig{
			KnownHostsPath: "ws_known_hosts",
			SSHTimeout:     time.Duration(timeoutSec) * time.Second,
			MaxSessions:    maxSessions,
			RateLimit:      rateLimit,
		},
	)
	return sshConnectionManager
}
