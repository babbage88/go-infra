package authapi

import (
	"context"
	"errors"
	"log/slog"

	"github.com/babbage88/go-infra/auth/hashing"
	"github.com/babbage88/go-infra/auth/jwt_auth"
	"github.com/babbage88/go-infra/database/db_access"
	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/utils/env_helper"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserAuthService struct {
	DbConn *pgxpool.Pool
	Envars *env_helper.EnvVars
}

type UserAuth interface {
	VerifyUserPassword(connPool *pgxpool.Pool) bool
	HashUserPassword()
	NewLoginRequest(username string, password string, isHashed bool) *UserLoginResponse
	CreateNewToken(userid string, role string, email string) (db_access.AuthTokenDao, error)
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

	if !qry.Enabled {
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

func (ua *UserAuthService) NewLoginRequest(username string, password string, isHashed bool) *UserLoginResponse {
	userloginReq := &UserLoginRequest{UserName: username, Password: password, IsHashed: isHashed}
	response := userloginReq.Login(ua.DbConn)

	return &response
}

func (ua *UserAuthService) CreateNewToken(userid int32, role string, email string) (db_access.AuthTokenDao, error) {
	token, err := jwt_auth.CreateToken(ua.Envars, userid, role, email)

	params := infra_db_pg.InsertAuthTokenParams{
		UserID:     pgtype.Int4{Int32: token.UserID, Valid: true},
		Token:      pgtype.Text{String: token.Token, Valid: true},
		Expiration: pgtype.Timestamp{Time: token.Expiration, Valid: true}}

	queries := infra_db_pg.New(ua.DbConn)
	queries.InsertAuthToken(context.Background(), params)

	return token, err
}
