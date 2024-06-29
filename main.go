package main

import (
	"database/sql"
	"log/slog"

	"github.com/babbage88/go-infra/auth/hashing"
	infra_db "github.com/babbage88/go-infra/database/infra_db"
	db_models "github.com/babbage88/go-infra/database/models"
	env_helper "github.com/babbage88/go-infra/utils/env_helper"
)

func createTestUserInstance(username string, password string, email string, tokens []string) db_models.User {
	hashedpw, err := hashing.HashPassword(password)
	if err != nil {
		slog.Error("Error hashing password", slog.String("Error", err.Error()))
	}

	testuser := db_models.User{
		Username:  username,
		Password:  hashedpw,
		Email:     email,
		ApiTokens: tokens,
	}

	return testuser
}

func initializeDbConn() *sql.DB {
	var db_pw = env_helper.NewDotEnvSource(env_helper.WithVarName("DB_PW")).GetEnvVarValue()
	dbConn := infra_db.NewDatabaseConnection(infra_db.WithDbHost("10.0.0.92"), infra_db.WithDbPassword(db_pw))

	db, _ := infra_db.InitializeDbConnection(dbConn)

	return db
}

func main() {

	//srvport := flag.String("srvadr", ":8993", "Address and port that http server will listed on. :8993 is default")
	//flag.Parse()
	//api_aerver.StartWebApiServer(db, srvport)

	db := initializeDbConn()
	var tokens []string

	tokens = append(tokens, "123456789")

	testuser := createTestUserInstance("testuser", "testpw", "test@trahan.dev", tokens)

	infra_db.InsertOrUpdateUser(db, &testuser)

	defer func() {
		if err := infra_db.CloseDbConnection(); err != nil {
			slog.Error("Failed to close the database connection: %v", err)
		}
	}()

}
