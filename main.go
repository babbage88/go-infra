package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"

	"github.com/babbage88/go-infra/auth/hashing"
	jwt_auth "github.com/babbage88/go-infra/auth/tokens"
	infra_db "github.com/babbage88/go-infra/database/infra_db"
	db_models "github.com/babbage88/go-infra/database/models"
	env_helper "github.com/babbage88/go-infra/utils/env_helper"
	"github.com/babbage88/go-infra/utils/test"
	"github.com/babbage88/go-infra/webapi/api_server"
)

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
	db_pw := envars.GetVarMapValue("DB_PW")
	db_host := envars.GetVarMapValue("DB_HOST")
	dbConn := infra_db.NewDatabaseConnection(infra_db.WithDbHost(db_host), infra_db.WithDbPassword(db_pw))

	db, _ := infra_db.InitializeDbConnection(dbConn)

	return db
}

func testUserDb(envars *env_helper.EnvVars, db *sql.DB) {
	testuser, _ := test.CreateTestUserInstance("jt", "testpw", "jt@trahan.dev", "admin")
	test.CreateUserDb(db, &testuser)

	user, _ := test.GetDbUserByUsername(db, testuser.Username)

	verify_pw := hashing.VerifyPassword("testpw", user.Password)

	if verify_pw {
		slog.Info("Password is verified for User: %s", slog.String("UserName", user.Username))
		slog.Info("Generating AuthToken for UserId", slog.String("UserId", fmt.Sprint(user.Id)))

		token, err := jwt_auth.CreateTokenanAddToDb(envars, db, user.Id, user.Role, user.Email)
		if err != nil {
			slog.Error("Error Generating JWT AuthToken", slog.String("Error", err.Error()))
		}

		fmt.Println(token.Token)
		jwt_auth.VerifyToken(token.Token)
	}

	if !verify_pw {
		fmt.Printf("Could not Verify Passworf for User: %s \n", user.Username)
	}

}

func main() {

	srvport := flag.String("srvadr", ":8993", "Address and port that http server will listed on. :8993 is default")
	hostEnvironment := flag.String("envfile", ".env", "Path to .env file to load Environment Variables.")
	flag.Parse()
	envars := env_helper.NewDotEnvSource(env_helper.WithDotEnvFileName(*hostEnvironment))
	envars.ParseEnvVariables()

	db := initializeDbConn(envars)
	testUserDb(envars, db)
	api_server.StartWebApiServer(envars, db, srvport)

	defer func() {
		if err := infra_db.CloseDbConnection(); err != nil {
			slog.Error("Failed to close the database connection: %v", err)
		}
	}()

}
