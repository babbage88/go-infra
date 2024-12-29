package main

import (
	"context"
	"os"

	"log/slog"

	"github.com/jackc/pgx/v5"
)

func CreateDbIfNotExist() {
	// Replace with your PostgreSQL connection details
	connStr := os.Getenv("DATABSE_BOOTSTRAP_URL")
	newDbName := os.Getenv("NEW_INFRA_DB")
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		slog.Error("Unable to connect to the database: ", slog.String("Error", err.Error()))
	}
	defer conn.Close(context.Background())

	// SQL query
	query := `
DO
$do$
BEGIN
   IF EXISTS (SELECT FROM pg_database WHERE datname = $1) THEN
      RAISE NOTICE 'Database already exists';  -- optional
   ELSE
    PERFORM dblink_exec('dbname=' || current_database(), 'CREATE DATABASE ' || quote_ident($1));

   END IF;
END
$do$;`

	// Execute the query
	_, err = conn.Exec(context.Background(), query, newDbName)
	if err != nil {
		slog.Error("Error executing SQL query: ", slog.String("Error", err.Error()))
	}

	slog.Info("Query executed successfully.")
}
