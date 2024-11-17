package test

import (
	"log"
	"log/slog"

	"github.com/babbage88/go-infra/auth/hashing"
	"github.com/babbage88/go-infra/database/db_access"
	"github.com/babbage88/go-infra/webapi/authapi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCreateNewUser(connPool *pgxpool.Pool, username string, pw string, email string, role string) (db_access.UserDao, error) {
	hashed_pw, err := hashing.HashPassword(pw)
	if err != nil {
		slog.Error("Error hashing password %s", err)
	}
	newuser, err := db_access.CreateUserQuery(connPool, username, hashed_pw, email, role)

	if err != nil {
		log.Fatalf("Error creating user %s", err)
	}
	return newuser, err
}

func TestUserLogin(connPool *pgxpool.Pool, username string, password string) authapi.UserLoginResponse {
	loginReq := authapi.UserLoginRequest{UserName: username, Password: password, IsHashed: false}
	response := loginReq.Login(connPool)
	if response.Result.Error != nil {
		slog.Error("Error during Login attempt")
	}
	return response
}
