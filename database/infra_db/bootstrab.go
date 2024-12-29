package infra_db

import (
	"context"
	"os"

	"log/slog"

	"github.com/babbage88/go-infra/internal/pretty"
	"github.com/jackc/pgx/v5"
)

func CreateDbIfNotExist() error {
	// Replace with your PostgreSQL connection details
	connStr := os.Getenv("DATABSE_BOOTSTRAP_URL")
	newDbName := os.Getenv("NEW_INFRA_DB")
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		slog.Error("Unable to connect to the database: ", slog.String("Error", err.Error()))
		return err
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
		return err
	}

	slog.Info("Query executed successfully.")
	return err
}

func main() {
	err := CreateDbIfNotExist()
	if err != nil {
		pretty.PrintErrorf("%s", err.Error())
		slog.Error("Error createing Database", slog.String("error", err.Error()))
	}
	pretty.Print("New database created successfully")
	slog.Info("New database created successfully")
}
