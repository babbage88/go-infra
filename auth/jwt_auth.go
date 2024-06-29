package jwt_auth

import (
	"fmt"
	"time"

	env_helper "github.com/babbage88/go-infra/utils/env_helper"
	"github.com/golang-jwt/jwt/v5"
)

var jwtkeydotEnv = env_helper.NewDotEnvSource().GetEnvVarValue()
var jwtKey = []byte(jwtkeydotEnv)

func createToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
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
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func VerifyToken(tokenString string) error {
	return verifyToken(tokenString)
}
