package jwt_auth

import (
	"fmt"
	"log/slog"
	"time"

	db_models "github.com/babbage88/go-infra/database/models"
	env_helper "github.com/babbage88/go-infra/utils/env_helper"
	"github.com/golang-jwt/jwt/v5"
)

var jwtkeydotEnv = env_helper.NewDotEnvSource().GetEnvVarValue()
var jwtKey = []byte(jwtkeydotEnv)

func createToken(userid int64) (db_models.AuthToken, error) {
	expire_time := time.Now().Add(time.Hour * 24)
	var retval db_models.AuthToken

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userid": userid,
			"exp":    expire_time.Unix(),
		})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {

		slog.Error("Error creating signed jwt token", slog.String("Error", err.Error()))
		return retval, err
	}

	retval = db_models.AuthToken{
		UserId:     userid,
		Expiration: expire_time,
		Token:      tokenString,
	}

	return retval, nil
}

func CreateToken(userid int64) (db_models.AuthToken, error) {
	return createToken(userid)
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
