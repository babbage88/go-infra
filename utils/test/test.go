package test

import (
	"fmt"
	"log/slog"

	"github.com/babbage88/go-infra/webapi/authapi"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestUserLogin(connPool *pgxpool.Pool, username string, password string) authapi.UserLoginResponse {
	loginReq := authapi.UserLoginRequest{UserName: username, Password: password}
	response := loginReq.Login(connPool)
	if response.Result.Error != nil {
		slog.Error("Error during Login attempt")
	}
	slog.Info("Login success:",
		slog.String("ID", fmt.Sprintf("%d", response.UserInfo.Id)),
		slog.String("Username", response.UserInfo.UserName),
		slog.String("Enabled", fmt.Sprintf("%t", response.UserInfo.Enabled)))
	return response
}
