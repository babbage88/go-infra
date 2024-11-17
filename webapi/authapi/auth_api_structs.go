package authapi

import (
	"context"
	"errors"
	"log/slog"

	"github.com/babbage88/go-infra/auth/hashing"
	"github.com/babbage88/go-infra/database/db_access"
	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ParsedCertbotOutput struct {
	CertificateInfo string `json:"certificateInfo"`
	Warnings        string `json:"warnings"`
	DebugLog        string `json:"debugLog"`
}

type LoginResult struct {
	Success         bool  `json:"success"`
	Error           error `json:"error"`
	UserNameMatches bool  `json:"username_matches"`
	PasswordValid   bool  `json:"password_valid"`
	UserEnabled     bool  `json:"enabled"`
}

type UserLoginRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
	IsHashed bool   `json:"isHashed"`
}

type UserLoginResponse struct {
	Result   LoginResult        `json:"result"`
	UserInfo *db_access.UserDao `json:"UserDao"`
}

type LoginActions interface {
	VerifyUserPassword(connPool *pgxpool.Pool) bool
	HashUserPassword()
	Login(connPool *pgxpool.Pool) UserLoginResponse
}

func (request *UserLoginRequest) HashUserPassword() {
	if !request.IsHashed {
		pw, err := hashing.HashPassword(request.Password)
		if err != nil {
			slog.Error("Error hashing password for user", slog.String("UserName", request.UserName))
			request.Password = pw
			request.IsHashed = true
		}
		request.Password = pw
	}
}

func (request *UserLoginRequest) Login(connPool *pgxpool.Pool) UserLoginResponse {
	var response UserLoginResponse
	var result LoginResult
	username := pgtype.Text{String: request.UserName, Valid: true}

	queries := infra_db_pg.New(connPool)
	qry, err := queries.GetUserLogin(context.Background(), username)
	result.PasswordValid = hashing.VerifyPassword(request.Password, qry.Password.String)

	if err != nil {
		slog.Error("Error querying database for user", slog.String("UserName", request.UserName))
	}

	if !result.PasswordValid {
		slog.Error("Supplied password does not match the password stored in database", slog.String("User", request.UserName))
		result.Success = false
		result.Error = errors.New("Password does not match.")
		result.UserEnabled = qry.Enabled
		response.Result = result
		return response
	}

	if qry.Enabled == false {
		slog.Error("User is disabled", slog.String("User", request.UserName))
		result.Success = false
		result.UserEnabled = qry.Enabled
		result.Error = errors.New("User is diabled.")
		response.Result = result
		return response
	}
	slog.Info("Login was Successful")
	result.Success = true
	result.Error = nil
	result.UserNameMatches = true

	response.Result = result
	response.UserInfo.ParseUserRowFromDb(qry)

	return response
}
