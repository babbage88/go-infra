package jwt_auth

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/babbage88/go-infra/services"
	env_helper "github.com/babbage88/go-infra/utils/env_helper"
	"github.com/golang-jwt/jwt/v5"
)

var jwtkeydotEnv = env_helper.NewDotEnvSource().GetEnvVarValue()
var jwtKey = []byte(jwtkeydotEnv)

type MyJWTClaims struct {
	*jwt.RegisteredClaims
	UserInfo interface{}
}

func createTokenString(envars *env_helper.EnvVars, sub string, userInfo interface{}) (string, time.Time, error) {
	var expire_minutes, err = envars.ParseEnvVarInt64("EXPIRATION_MINUTES")
	var jwt_algo = envars.GetVarMapValue("JWT_ALGORITHM")

	if err != nil {
		slog.Error("Error Parsing int64 from .env EXPIRATION_MINUTES, setting value to 60.", slog.String("Error", err.Error()))
		expire_minutes = 60
	}
	token := jwt.New(jwt.GetSigningMethod(jwt_algo))
	exp := time.Now().Add(time.Minute * time.Duration(expire_minutes))
	token.Claims = &MyJWTClaims{
		&jwt.RegisteredClaims{
			// Set the userid and expiration as the standard claim.
			ExpiresAt: jwt.NewNumericDate(exp),
			Subject:   sub,
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

func createToken(envars *env_helper.EnvVars, userid int32, role string, email string) (services.AuthTokenDao, error) {

	var retval services.AuthTokenDao
	userInfo := map[string]interface{}{
		"role":  role,
		"email": email,
	}

	tokenString, expire_time, err := createTokenString(envars, fmt.Sprint(userid), userInfo)
	if err != nil {
		slog.Error("Error creating signed jwt token", slog.String("Error", err.Error()))
		return retval, err
	}

	retval = services.AuthTokenDao{
		UserID:     userid,
		Expiration: expire_time,
		Token:      tokenString,
	}

	return retval, nil
}

func CreateToken(envars *env_helper.EnvVars, userid int32, role string, email string) (services.AuthTokenDao, error) {
	return createToken(envars, userid, role, email)
}

func CreateTokenanAddToDb(envars *env_helper.EnvVars, db *sql.DB, userid int32, role string, email string) (services.AuthTokenDao, error) {
	token, err := createToken(envars, userid, role, email)
	if err != nil {
		slog.Error("Error creating signed token", slog.String("Error", err.Error()))
	}

	//infra_db.InsertAuthToken(db, &token)

	return token, nil
}

func verifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		slog.Error("Failed parsing jwt token", slog.String("Error", err.Error()))
		return err
	}

	if !token.Valid {
		slog.Error("Token is not valid.", slog.String("Error", err.Error()))
		return fmt.Errorf("invalid token")
	}

	return nil
}

func VerifyToken(tokenString string) error {
	return verifyToken(tokenString)
}
