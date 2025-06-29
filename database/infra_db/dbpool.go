package infra_db

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DbInfo struct {
	DbName        string `json:"dbName" yaml:"db_name" db:"DbName"`
	DbUser        string `json:"dbUser" yaml:"db_user" db:"DbUser"`
	ServerAddress string `json:"serverAddress" yaml:"server_address" db:"ServerAddress"`
	ServerPort    uint16 `json:"serverPort" yaml:"server_port" db:"ServerPort"`
	ClientAddress string `json:"clientAddress" yaml:"client_address" db:"ClientAddress"`
	ClientPort    uint16 `json:"clientPort" yaml:"client_port" db:"ClientPort"`
}

func (db *DbInfo) ServerPortString() string {
	srvPort := fmt.Sprintf("%d", db.ServerPort)
	return srvPort
}

func (db *DbInfo) ClientPortString() string {
	clientPort := fmt.Sprintf("%d", db.ClientPort)
	return clientPort
}

func getDbInfo(connection *pgxpool.Conn) (*DbInfo, error) {
	dbInfo := &DbInfo{}
	var getDbNameQuery = `SELECT current_database() as "DbName", 
	session_user AS "DbUser", 
	inet_server_addr()::text as "ServerAddress", 
	inet_server_port() as "ServerPort", 
	inet_client_addr()::text "ClientAddress", 
	inet_client_port() ClientPort;`
	row := connection.QueryRow(context.Background(), getDbNameQuery)
	err := row.Scan(&dbInfo.DbName, &dbInfo.DbUser, &dbInfo.ServerAddress, &dbInfo.ServerPort, &dbInfo.ClientAddress, &dbInfo.ClientPort)
	if err != nil {
		slog.Error("Error retrieving DbName and DbUser from database", slog.String("error", err.Error()))
		return dbInfo, err
	}
	return dbInfo, err
}

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
		slog.Debug("Connection Aquired!")
		return true
	}

	dbConfig.AfterRelease = func(c *pgx.Conn) bool {
		return true
	}

	dbConfig.BeforeClose = func(c *pgx.Conn) {
		slog.Debug("Close Conneciion Pool!")
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

	dbInfo, infoErr := getDbInfo(connection)
	if infoErr != nil {
		slog.Error("error retrieving db info", slog.String("error", err.Error()))
	}

	os.Setenv("DB_NAME", dbInfo.DbName)
	os.Setenv("DB_USER", dbInfo.DbUser)
	os.Setenv("DB_ADDRESS", dbInfo.ServerAddress)
	os.Setenv("DB_PORT", dbInfo.ServerPortString())

	slog.Info("Connected to the database!", "Database", os.Getenv("DB_NAME"), "User", dbInfo.DbUser, "Address", dbInfo.ServerAddress, "Port", dbInfo.ServerPortString())

	return connPool
}
