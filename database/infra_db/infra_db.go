package infra_db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"sync"

	type_helper "github.com/babbage88/go-infra/utils/type_helper"
)

type DatabaseConnection struct {
	DbHost     string `json:"dbHost"`
	DbPort     int32  `json:"dbPort"`
	DbUser     string `json:"dbUser"`
	DbPassword string `json:"dbPassword"`
	DbName     string `json:"database"`
}

type DatabaseConnectionOptions func(*DatabaseConnection)

// Global db instance
var (
	db     *sql.DB
	dbOnce sync.Once
	dbErr  error
)

func InitializeDbConnection(dbConn *DatabaseConnection) (*sql.DB, error) {
	dbOnce.Do(func() {
		// Connect to the PostgreSQL database
		slog.Info("Connecting to database: %s on server %s", dbConn.DbName, dbConn.DbHost)

		psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbConn.DbHost, type_helper.String(dbConn.DbPort), dbConn.DbUser, dbConn.DbPassword, dbConn.DbName)

		db, dbErr = sql.Open("postgres", psqlInfo)
		if dbErr != nil {
			slog.Error("Error connecting to the database", slog.String("Error", dbErr.Error()))
		}
	})
	return db, dbErr
}

func CloseDbConnection() error {
	if db != nil {
		slog.Info("Closing DB Connection")
		return db.Close()
	}
	return nil
}

func NewDatabaseConnection(opts ...DatabaseConnectionOptions) *DatabaseConnection {
	const (
		dbHost = "localhost"
		dbPort = 5432
		dbUser = "postgres"
		dbName = "go-infra"
	)
	dbPassword := os.Getenv("DB_PASSWORD")

	db := &DatabaseConnection{
		DbHost:     dbHost,
		DbPort:     dbPort,
		DbUser:     dbUser,
		DbPassword: dbPassword,
		DbName:     dbName,
	}

	for _, opt := range opts {
		opt(db)
	}

	return db
}

func WithDbHost(DbHostname string) DatabaseConnectionOptions {
	return func(c *DatabaseConnection) {
		c.DbHost = DbHostname
	}
}

func WithDbPort(dbPort int32) DatabaseConnectionOptions {
	return func(c *DatabaseConnection) {
		c.DbPort = dbPort
	}
}

func WithDbUser(dbUser string) DatabaseConnectionOptions {
	return func(c *DatabaseConnection) {
		c.DbUser = dbUser
	}
}

func WithDbPassword(dbPassword string) DatabaseConnectionOptions {
	return func(c *DatabaseConnection) {
		c.DbPassword = dbPassword
	}
}

func WithDbName(dbName string) DatabaseConnectionOptions {
	return func(c *DatabaseConnection) {
		c.DbName = dbName
	}
}
