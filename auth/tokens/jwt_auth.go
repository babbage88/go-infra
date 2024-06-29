package jwt_auth

import (
	"fmt"
	"log/slog"
	"time"

	env_helper "github.com/babbage88/go-infra/utils/env_helper"
	"github.com/golang-jwt/jwt/v5"
)

var jwtkeydotEnv = env_helper.NewDotEnvSource().GetEnvVarValue()
var jwtKey = []byte(jwtkeydotEnv)

func createToken(userid string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userid": userid,
			"exp":    time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {

		slog.Error("Error creating jwt token", slog.String("Error", err.Error()))
		return "", err
	}

	return tokenString, nil
}

func CreateToken(username string) (string, error) {
	return createToken(username)
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
