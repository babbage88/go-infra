package authapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/internal/type_helper"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LocalAuthService struct {
	DbConn *pgxpool.Pool `json:"dbConn"`
}

func NewLoginRequest(username string, password string, isHashed bool) *UserLoginRequest {
	return &UserLoginRequest{UserName: username, Password: password}
}

func (ua *LocalAuthService) VerifyUser(userid uuid.UUID) bool {
	queries := infra_db_pg.New(ua.DbConn)
	qry, err := queries.GetUserById(context.Background(), userid)
	if err != nil {
		slog.Error("Error querying database for user", slog.String("Error", err.Error()), slog.String("UserName", fmt.Sprint(userid)))
	}
	return qry.Enabled
}

func (us *LocalAuthService) VerifyUserPermission(ueid uuid.UUID, permissionName string) (bool, error) {
	params := infra_db_pg.VerifyUserPermissionByIdParams{
		UserId:     pgtype.UUID{Bytes: ueid, Valid: true},
		Permission: pgtype.Text{String: permissionName, Valid: true},
	}
	queries := infra_db_pg.New(us.DbConn)
	qry, err := queries.VerifyUserPermissionById(context.Background(), params)
	if err != nil {
		slog.Error("error verifying user permissions", slog.String("error", err.Error()))
		return false, err
	}
	return qry, err
}

func (us *LocalAuthService) VerifyUserPermissionByRole(roleId uuid.UUID, permissionName string) (bool, error) {
	params := infra_db_pg.VerifyUserPermissionByRoleIdParams{
		RoleId:     roleId,
		Permission: pgtype.Text{String: permissionName, Valid: true},
	}

	queries := infra_db_pg.New(us.DbConn)
	qry, err := queries.VerifyUserPermissionByRoleId(context.Background(), params)
	if err != nil {
		slog.Error("error verifying user permissions", slog.String("error", err.Error()))
		return false, err
	}
	return qry, err
}

func (a *LocalAuthService) CreateAuthTokenOnLogin(id uuid.UUID, roleIds uuid.UUIDs, email string) (AuthToken, error) {
	tokens := AuthToken{UserID: id}
	var expireMinutes int
	envExp := os.Getenv("EXPIRATION_MINUTES")
	expireMinutesInt, err := type_helper.ParseIntegerFromString[int](envExp)
	if err != nil {
		slog.Error("Error parsing JWT Expiration minutes from env var EXPIRATION_MINUTS")
		expireMinutes = int(15)
	}
	expireMinutes = expireMinutesInt
	signingMethod := getJwtSigningMenthodFromEnv()
	expTime := time.Now().Add(time.Minute * time.Duration(expireMinutes))
	tokens.Expiration = expTime

	// Create access accessToken
	accessToken, err := NewAccessTokenWithExp(id, roleIds, email, signingMethod, expTime)

	if err != nil {
		return tokens, err
	}
	tokens.Token = accessToken

	refreshToken, err := NewRefreshToken(id, signingMethod)

	if err != nil {
		return tokens, err
	}

	tokens.RefreshToken = refreshToken
	return tokens, nil
}

func (a *LocalAuthService) Login(loginReq *UserLoginRequest) UserLoginResponse {
	var response UserLoginResponse
	var result LoginResult
	username := pgtype.Text{String: loginReq.UserName, Valid: true}

	queries := infra_db_pg.New(a.DbConn)
	qry, err := queries.GetUserLogin(context.Background(), username)
	result.PasswordValid = VerifyPassword(loginReq.Password, qry.Password.String)

	if err != nil {
		slog.Error("Error querying database for user", slog.String("UserName", loginReq.UserName))
	}

	if !result.PasswordValid {
		slog.Error("Supplied password does not match the password stored in database", slog.String("User", loginReq.UserName))
		result.Success = false
		result.Error = errors.New("password does not match")
		result.UserEnabled = qry.Enabled
		response.Result = result
		return response
	}

	if !qry.Enabled {
		slog.Error("User is disabled", slog.String("User", loginReq.UserName))
		result.Success = false
		result.UserEnabled = qry.Enabled
		result.Error = errors.New("user is diabled")
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

func (a *LocalAuthService) VerifyToken(tokenString string) error {
	jwtKey := []byte(os.Getenv("JWT_KEY"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func (ua *LocalAuthService) ParseAccessToken(accessToken string) *InfraJWTClaim {
	jwtKey := os.Getenv("JWT_KEY")
	parsedAccessToken, _ := jwt.ParseWithClaims(accessToken, &InfraJWTClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})

	return parsedAccessToken.Claims.(*InfraJWTClaim)
}

func (ua *LocalAuthService) VerifyUserRolesForPermission(roleIds uuid.UUIDs, permissionName string) (bool, error) {
	var lastError error // Store any encountered errors for logging or debugging

	for _, roleId := range roleIds {
		hasPermission, err := ua.VerifyUserPermissionByRole(roleId, permissionName)
		if err != nil {
			// Save the error but continue checking other roles
			slog.Error("Error encountered while verifying permissions", slog.String("roleId", roleId.String()), slog.String("error", err.Error()))
			lastError = err
			continue
		}
		if hasPermission {
			return true, err
		}
	}

	if lastError != nil {
		// Log the error for debugging purposes
		slog.Error("Error occurred while verifying permissions for roles", slog.String("Error", lastError.Error()))
	}
	// Return false if no roles grant the permission
	return false, lastError
}

func getJwtSigningMenthodFromEnv() jwt.SigningMethod {
	jwt_algo := os.Getenv("JWT_ALGORITHM")
	if len(jwt_algo) < 1 {
		slog.Info("No JWT_ALGORITHM set, defaulting to HS256")
		return jwt.SigningMethodHS256
	}

	signingMethod := jwt.GetSigningMethod(jwt_algo)
	return signingMethod
}
