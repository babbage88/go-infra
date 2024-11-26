package authapi

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/babbage88/go-infra/utils/type_helper"
	"github.com/golang-jwt/jwt/v5"
)

func NewInfraJWTClaims(userid int32, userInfo interface{}) (InfraJWTClaim, error) {
	env_expire_minutes := os.Getenv("EXPIRATION_MINUTES")
	expire_minutes, err := type_helper.ParseInt64(env_expire_minutes)
	if err != nil {
		slog.Error("Error Parsing int64 from .env EXPIRATION_MINUTES, setting value to 60.", slog.String("Error", err.Error()))
		expire_minutes = 60
	}

	exp := time.Now().Add(time.Minute * time.Duration(expire_minutes))
	retVal := &InfraJWTClaim{
		&jwt.RegisteredClaims{
			// Set the userid and expiration as the standard claim.
			Issuer:    "goinfra",
			ExpiresAt: jwt.NewNumericDate(exp),
			Subject:   fmt.Sprint(userid),
		},
		// UserInfo passed from caller as map[string]string
		userInfo,
	}
	return *retVal, nil
}

func NewAccessToken(claims InfraJWTClaim, algo *jwt.SigningMethod) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return accessToken.SignedString([]byte(os.Getenv("JWT_KEY")))
}

func NewRefreshToken(claims jwt.RegisteredClaims) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return refreshToken.SignedString([]byte(os.Getenv("JWT_KEY")))
}

func ParseAccessToken(accessToken string) *InfraJWTClaim {
	parsedAccessToken, _ := jwt.ParseWithClaims(accessToken, &InfraJWTClaim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_KEY")), nil
	})

	return parsedAccessToken.Claims.(*InfraJWTClaim)
}

func ParseRefreshToken(refreshToken string) *jwt.RegisteredClaims {
	parsedRefreshToken, _ := jwt.ParseWithClaims(refreshToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_KEY")), nil
	})

	return parsedRefreshToken.Claims.(*jwt.RegisteredClaims)
}
