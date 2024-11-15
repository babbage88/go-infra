package test

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func pgxPoolConfig() *pgxpool.Config {
	const defaultMaxConns = int32(4)
	const defaultMinConns = int32(0)
	const defaultMaxConnLifetime = time.Hour
	const defaultMaxConnIdleTime = time.Minute * 30
	const defaultHealthCheckPeriod = time.Minute
	const defaultConnectTimeout = time.Second * 5
	connString := os.Getenv("DATABASE_URL")

	dbConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		slog.Error("Failed to create a config, error: ", "Error", err)
	}

	dbConfig.MaxConns = defaultMaxConns
	dbConfig.MinConns = defaultMinConns
	dbConfig.MaxConnLifetime = defaultMaxConnLifetime
	dbConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	dbConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout
	dbConfig.BeforeAcquire = func(ctx context.Context, c *pgx.Conn) bool {
		slog.Info("Before acquiring the connection pool to the database!!")
		return true
	}

	dbConfig.AfterRelease = func(c *pgx.Conn) bool {
		slog.Info("After releasing the connection pool to the database!!")
		return true
	}

	dbConfig.BeforeClose = func(c *pgx.Conn) {
		log.Println("Closed the connection pool to the database!!")
	}

	return dbConfig

}

func TestCreateUserQuery(username string, hashed_pw string) (infra_db_pg.User, error) {
	// Create database connection
	connPool, err := pgxpool.NewWithConfig(context.Background(), pgxPoolConfig())
	if err != nil {
		slog.Error("Error while creating connection to the database!!", "Error", err)
	}

	connection, err := connPool.Acquire(context.Background())
	if err != nil {
		slog.Error("Error while acquiring connection from the database pool!!", "Error", err)
	}
	defer connection.Release()

	err = connection.Ping(context.Background())
	if err != nil {
		slog.Error("Could not ping database")
	}

	slog.Info("Connected to the database!!", "Database", os.Getenv("DATABASE_URL"))

	// Set up parameters for the new user
	params := infra_db_pg.CreateUserParams{
		Username: pgtype.Text{String: "johndoe", Valid: true},
		Password: pgtype.Text{String: "password123", Valid: true},
		Email:    pgtype.Text{String: "johndoe@example.com", Valid: true},
		Role:     pgtype.Text{String: "user", Valid: true},
	}

	queries := infra_db_pg.New(connPool)
	newUser, err := queries.CreateUser(context.Background(), params)
	return newUser, err
}
