package main

import (
	"database/sql"
	"log/slog"

	infra_db "github.com/babbage88/go-infra/database/infra_db"
	env_helper "github.com/babbage88/go-infra/utils/env_helper"
)

func main() {
	var db_pw = env_helper.NewDotEnvSource(env_helper.WithVarName("DB_PW")).GetEnvVarValue()
	dbConn := infra_db.NewDatabaseConnection(infra_db.WithDbHost("10.0.0.92"), infra_db.WithDbPassword(db_pw))
	db, _ := infra_db.InitializeDbConnection(dbConn)

	// Create tables
	createTables(db)

	// Perform CRUD operations here
}

func createTables(db *sql.DB) {
	hostServerTable := `CREATE TABLE IF NOT EXISTS host_servers (
		id SERIAL PRIMARY KEY,
		hostname VARCHAR(255) NOT NULL,
		ip_address INET NOT NULL,
		username VARCHAR(255),
		public_ssh_keyname VARCHAR(255),
		hosted_domains VARCHAR(255)[],
		ssl_key_path VARCHAR(4096),
		is_container_host BOOLEAN,
		is_vm_host BOOLEAN,
		is_virtual_machine BOOLEAN,
		is_db_host BOOLEAN,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_modified TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT unique_hostname_ip UNIQUE (hostname, ip_address)
	);`

	userTable := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255),
		password TEXT,
		email VARCHAR(255),
		role VARCHAR(255),
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_modified TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT unique_username UNIQUE (username),
		CONSTRAINT unique_email UNIQUE (email)
	);`

	authTokenTable := `CREATE TABLE IF NOT EXISTS auth_tokens (
		id SERIAL PRIMARY KEY,
		user_id INTEGER REFERENCES users(id),
		token TEXT,
		expiration TIMESTAMP,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		last_modified TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
	// Execute SQL statements
	_, err := db.Exec(hostServerTable)
	if err != nil {
		slog.Error("Error Creating Host Table", slog.String("Error", err.Error()))
	}

	_, err = db.Exec(userTable)
	if err != nil {
		slog.Error("Error Creating Host Table", slog.String("Error", err.Error()))
	}

	_, err = db.Exec(authTokenTable)
	if err != nil {
		slog.Error("Error Creating Host Table", slog.String("Error", err.Error()))
	}
	slog.Info("Tables have been created.")
}
