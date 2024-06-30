package test

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/babbage88/go-infra/auth/hashing"
	infra_db "github.com/babbage88/go-infra/database/infra_db"
	db_models "github.com/babbage88/go-infra/database/models"
	env_helper "github.com/babbage88/go-infra/utils/env_helper"
)

func CreateTestUserInstance(username string, password string, email string, role string) (db_models.User, error) {
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

	return testuser, nil
}

func InitializeDbConn() (*sql.DB, error) {
	var db_host = env_helper.NewDotEnvSource(env_helper.WithVarName("DB_HOST")).GetEnvVarValue()
	var db_pw = env_helper.NewDotEnvSource(env_helper.WithVarName("DB_PW")).GetEnvVarValue()
	var db_user = env_helper.NewDotEnvSource(env_helper.WithVarName("DB_USER")).GetEnvVarValue()
	var db_port, _ = env_helper.NewDotEnvSource(env_helper.WithVarName("DB_PORT")).ParseEnvVarInt32()

	dbConn := infra_db.NewDatabaseConnection(
		infra_db.WithDbHost(db_host),
		infra_db.WithDbPassword(db_pw),
		infra_db.WithDbUser(db_user),
		infra_db.WithDbPort(db_port),
	)

	db, err := infra_db.InitializeDbConnection(dbConn)
	if err != nil {
		slog.Error("Error initializing Database connection", slog.String("Error", err.Error()),
			slog.String("DB_HOST", dbConn.DbHost),
			slog.String("DB_USER", dbConn.DbUser),
			slog.String("DB_PORT", fmt.Sprint(dbConn.DbPort)))
	}

	return db, nil
}

func CreateUserDb(db *sql.DB, user *db_models.User) error {
	err := infra_db.InsertOrUpdateUser(db, user)
	if err != nil {
		slog.Error("Error adding or updating user in databse", slog.String("Error", err.Error()))
	}

	return err
}

func AddAuthTokenToDb(db *sql.DB, token *db_models.AuthToken) error {
	err := infra_db.InsertAuthToken(db, token)
	if err != nil {
		slog.Error("Error adding or updating AuthToken in databse", slog.String("Error", err.Error()))
	}

	return err
}

func AddHostToDb(db *sql.DB, host *db_models.HostServer) error {
	var host_slice []db_models.HostServer = make([]db_models.HostServer, 1)
	host_slice = append(host_slice, *host)

	err := infra_db.InsertOrUpdateHostServer(db, host_slice)
	if err != nil {
		slog.Error("Error adding or updating HostServer in databse", slog.String("Error", err.Error()))
	}

	return err
}

func AddHostsToDb(db *sql.DB, hosts []db_models.HostServer) error {
	err := infra_db.InsertOrUpdateHostServer(db, hosts)
	if err != nil {
		slog.Error("Error adding or updating HostServers in databse", slog.String("Error", err.Error()))
	}

	return err
}

func GetDbUserByUsername(db *sql.DB, username string) (*db_models.User, error) {
	user, err := infra_db.GetUserByUsername(db, username)
	if err != nil {
		slog.Error("Error retrieving user from databse", slog.String("Error", err.Error()))
	}

	return user, nil
}

func GetDbUserById(db *sql.DB, id int64) (*db_models.User, error) {
	user, err := infra_db.GetUserById(db, id)
	if err != nil {
		slog.Error("Error retrieving user from databse", slog.String("Error", err.Error()))
	}

	return user, nil
}

func GetDbAuthToken(db *sql.DB, tokenstr string) (*db_models.AuthToken, error) {
	token, err := infra_db.GetAuthTokenFromDb(db, tokenstr)
	if err != nil {
		slog.Error("Error retrieving AuthToken from databse", slog.String("Error", err.Error()))
	}

	return token, nil
}
