package authapi

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/babbage88/go-infra/auth/hashing"
	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/utils/env_helper"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserAuthService struct {
	DbConn *pgxpool.Pool
	Envars *env_helper.EnvVars
}

type UserAuth interface {
	VerifyUser(userid int32) bool
	RefreshAuthTokens(dbConn *pgxpool.Pool) error
	HashUserPassword()
	NewLoginRequest(username string, password string, isHashed bool) *UserLoginResponse
	CreateAuthToken(userid int32, role string, email string) (AuthToken, error)
	CreateSignedTokenString(sub string, userInfo interface{}) (string, time.Time, error)
	VerifyToken(tokenString string) error
}

func (ua *UserAuthService) VerifyUser(userid int32) bool {
	queries := infra_db_pg.New(ua.DbConn)
	qry, err := queries.GetUserById(context.Background(), userid)
	if err != nil {
		slog.Error("Error querying database for user", slog.String("Error", err.Error()), slog.String("UserName", fmt.Sprint(userid)))
	}
	return qry.Enabled
}

func (t *AuthToken) RefreshAuthTokens(dbConn *pgxpool.Pool) error {
	queries := infra_db_pg.New(dbConn)
	qry, err := queries.GetUserById(context.Background(), t.UserID)
	if err != nil {
		slog.Error("Error querying database for user", slog.String("Error", err.Error()), slog.String("UserName", fmt.Sprint(t.UserID)))
	}
	if !qry.Enabled {
		slog.Warn("User is not enabled", slog.String("UserID", fmt.Sprint(t.UserID)))
		return fmt.Errorf("User is not enabled")
	}
	userInfo := map[string]interface{}{
		"role":  qry.Role,
		"email": qry.Email,
	}
	jwt_algo := os.Getenv("JWT_ALGORITHM")
	signingMethod := jwt.GetSigningMethod(jwt_algo)

	claims, err := NewInfraJWTClaims(t.UserID, userInfo)
	newAccessToken, err := NewAccessToken(claims, &signingMethod)
	if err != nil {
		slog.Error("Error creating new auth token.", slog.String("Error", err.Error()))
		return err
	}

	t.Token = newAccessToken

	return nil
}

func (request *UserLoginRequest) HashUserPassword() {
	pw, err := hashing.HashPassword(request.Password)
	if err != nil {
		slog.Error("Error hashing password for user", slog.String("UserName", request.UserName))
		request.Password = pw
	}
	request.Password = pw
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

func (ua *UserAuthService) NewLoginRequest(username string, password string) *UserLoginResponse {
	userloginReq := &UserLoginRequest{UserName: username, Password: password}
	response := userloginReq.Login(ua.DbConn)

	return &response
}

func (t *AuthToken) CreateRefreshToken() {
	jwtKey := os.Getenv("JWT_KEY")
	//refreshEnvValue, err := type_helper.ParseInt64(os.Getenv("REFRESH_EXPIRATION_HOURS"))
	refreshExpiration := time.Now().Add(time.Hour * 48).Unix()
	refreshToken := jwt.New(jwt.SigningMethodHS256)

	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["sub"] = t.UserID
	rtClaims["exp"] = refreshExpiration

	rt, err := refreshToken.SignedString([]byte(jwtKey))
	if err != nil {
		slog.Error("Error signing refresh token", slog.String("Error", err.Error()))
	}

	t.RefreshToken = rt
}

func (ua *UserAuthService) CreateSignedAuthTokenString(sub string, role string, userInfo interface{}) (string, time.Time, error) {
	expire_minutes, err := ua.Envars.ParseEnvVarInt64("EXPIRATION_MINUTES")
	jwt_algo := ua.Envars.GetVarMapValue("JWT_ALGORITHM")
	jwtKey := []byte(ua.Envars.GetVarMapValue("JWT_KEY"))

	if err != nil {
		slog.Error("Error Parsing int64 from .env EXPIRATION_MINUTES, setting value to 60.", slog.String("Error", err.Error()))
		expire_minutes = 60
	}
	token := jwt.New(jwt.GetSigningMethod(jwt_algo))
	exp := time.Now().Add(time.Minute * time.Duration(expire_minutes))
	token.Claims = &InfraJWTClaim{
		&jwt.RegisteredClaims{
			// Set the userid and expiration as the standard claim.
			Issuer:    "goinfra",
			ExpiresAt: jwt.NewNumericDate(exp),
			Subject:   sub,
			Audience:  jwt.ClaimStrings{role},
		},
		// UserInfo passed from caller as map[string]string
		userInfo,
	}
	val, err := token.SignedString(jwtKey)

	if err != nil {
		return "", exp, err
	}
	return val, exp, nil
}

func (ua *UserAuthService) CreateAuthTokenOnLogin(userid int32, role string, email string) (AuthToken, error) {

	var retval AuthToken
	userInfo := map[string]interface{}{
		"role":  role,
		"email": email,
	}

	tokenString, expire_time, err := ua.CreateSignedAuthTokenString(fmt.Sprint(userid), role, userInfo)
	if err != nil {
		slog.Error("Error creating signed jwt token", slog.String("Error", err.Error()))
		return retval, err
	}

	retval = AuthToken{
		UserID:     userid,
		Expiration: expire_time,
		Token:      tokenString,
	}

	retval.CreateRefreshToken()

	return retval, nil
}

func (ua *UserAuthService) VerifyToken(tokenString string) error {
	jwtKey := []byte(ua.Envars.GetVarMapValue("JWT_KEY"))
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

func (ua *UserAuthService) ParseAccessToken(accessToken string) *InfraJWTClaim {
	jwtKey := ua.Envars.GetVarMapValue("JWT_KEY")
	parsedAccessToken, _ := jwt.ParseWithClaims(accessToken, &InfraJWTClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})

	return parsedAccessToken.Claims.(*InfraJWTClaim)
}
