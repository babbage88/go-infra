package main

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/babbage88/go-infra/auth/hashing"
	jwt_auth "github.com/babbage88/go-infra/auth/tokens"
	infra_db "github.com/babbage88/go-infra/database/infra_db"
	db_models "github.com/babbage88/go-infra/database/models"
	env_helper "github.com/babbage88/go-infra/utils/env_helper"
	"github.com/babbage88/go-infra/utils/test"
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

func initializeDbConn() *sql.DB {
	var db_pw = env_helper.NewDotEnvSource(env_helper.WithVarName("DB_PW")).GetEnvVarValue()
	dbConn := infra_db.NewDatabaseConnection(infra_db.WithDbHost("10.0.0.92"), infra_db.WithDbPassword(db_pw))

	db, _ := infra_db.InitializeDbConnection(dbConn)

	return db
}

func testUserDb(db *sql.DB) {
	testuser, _ := test.CreateTestUserInstance("jt", "testpw", "jt@trahan.dev", "admin")
	test.CreateUserDb(db, &testuser)

	user, _ := test.GetDbUserByUsername(db, testuser.Username)

	verify_pw := hashing.VerifyPassword("testpw", user.Password)

	if verify_pw {
		slog.Info("Password is verified for User: %s", slog.String("UserName", user.Username))
		slog.Info("Generating AuthToken for UserId", slog.String("UserId", fmt.Sprint(user.Id)))

		token, err := jwt_auth.CreateTokenanAddToDb(db, user.Id, user.Role, user.Email)
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

	//srvport := flag.String("srvadr", ":8993", "Address and port that http server will listed on. :8993 is default")
	//flag.Parse()
	//api_aerver.StartWebApiServer(db, srvport)

	db := initializeDbConn()
	testUserDb(db)

	defer func() {
		if err := infra_db.CloseDbConnection(); err != nil {
			slog.Error("Failed to close the database connection: %v", err)
		}
	}()

}
