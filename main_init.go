package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"

	"github.com/babbage88/go-infra/auth/hashing"
	infra_db "github.com/babbage88/go-infra/database/infra_db"
	db_models "github.com/babbage88/go-infra/database/models"
	env_helper "github.com/babbage88/go-infra/utils/env_helper"
	"github.com/babbage88/go-infra/webapi/api_server"
	"github.com/jackc/pgx/v5/pgxpool"
)

func initEnvironment(envPath string) *env_helper.EnvVars {
	env_helper.LoadEnvFile(*&envPath)
	envars := env_helper.NewDotEnvSource(env_helper.WithDotEnvFileName(*&envPath))
	fmt.Printf("EnVars file name: %s\n", envars.DotFileName)
	envars.ParseEnvVariables()
	return envars
}

func createTestUserInstance(username string, password string, email string, role string) db_models.User {
	hashedpw, err := hashing.HashPassword(password)
	if err != nil {
		slog.Error("Error hashing password", slog.String("Error", err.Error()))
	}

	testuser := db_models.User{
		Username: username,
		Password: hashedpw,
		Email:    email,
		Role:     role,
	}

	return testuser
}

func initializeDbConn(envars *env_helper.EnvVars) *sql.DB {
	db_name := envars.GetVarMapValue("DB_NAME")
	db_pw := envars.GetVarMapValue("DB_PW")
	db_user := envars.GetVarMapValue("DB_USER")
	db_host := envars.GetVarMapValue("DB_HOST")
	db_port, err := envars.ParseEnvVarInt32("DB_PORT")
	if err != nil {
		fmt.Errorf("Error Parsing DB_PORT from .env file", err)
	}
	dbConn := infra_db.NewDatabaseConnection(infra_db.WithDbHost(db_host),
		infra_db.WithDbPassword(db_pw),
		infra_db.WithDbUser(db_user),
		infra_db.WithDbPort(db_port),
		infra_db.WithDbName(db_name))

	db, _ := infra_db.InitializeDbConnection(dbConn)

	return db
}

func initPgConnPool() *pgxpool.Pool {
	connPool := infra_db.PgPoolInit()
	return connPool
}

func oldMain() {
	srvport := flag.String("srvadr", ":8993", "Address and port that http server will listed on. :8993 is default")
	hostEnvironment := flag.String("envfile", ".env", "Path to .env file to load Environment Variables.")
	version := flag.Bool("version", false, "Show the current version.")
	flag.Parse()

	if *version {
		showVersion()
		return
	}

	env_helper.LoadEnvFile(*hostEnvironment)
	envars := env_helper.NewDotEnvSource(env_helper.WithDotEnvFileName(*hostEnvironment))
	fmt.Printf("EnVars file name: %s\n", envars.DotFileName)
	envars.ParseEnvVariables()

	db := initializeDbConn(envars)

	defer func() {
		if err := infra_db.CloseDbConnection(); err != nil {
			slog.Error("Failed to close the database connection: ", "Error", err)
		}
	}()

	api_server.StartWebApiServer(envars, db, srvport)

}
