package infra_db

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func pgxPoolConfig() *pgxpool.Config {
	const defaultMaxConns = int32(8)
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
		slog.Info("Before Pg Connection Aquired!")
		return true
	}

	dbConfig.AfterRelease = func(c *pgx.Conn) bool {
		slog.Info("Conection Pool handle released!")
		return true
	}

	dbConfig.BeforeClose = func(c *pgx.Conn) {
		log.Println("Before Close Conneciion Pool!")
	}

	return dbConfig

}

func PgPoolInit() *pgxpool.Pool {
	// Create database connection
	connPool, err := pgxpool.NewWithConfig(context.Background(), pgxPoolConfig())
	if err != nil {
		slog.Error("Error while creating connection to the database!", "Error", err)
	}

	connection, err := connPool.Acquire(context.Background())
	if err != nil {
		slog.Error("Error while acquiring connection from the database pool!", "Error", err)
	}
	defer connection.Release()

	err = connection.Ping(context.Background())
	if err != nil {
		slog.Error("Could not ping database")
	}

	slog.Info("Connected to the database!", "Database", os.Getenv("DB_NAME"))

	return connPool
}
