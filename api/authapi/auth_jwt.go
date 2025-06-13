package authapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/internal/type_helper"
	"github.com/babbage88/go-infra/services/user_crud_svc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func NewAccessTokenWithExp(id uuid.UUID, roleIds uuid.UUIDs, email string, signingMethod jwt.SigningMethod, expTime time.Time) (string, error) {
	// Create token
	token := jwt.New(signingMethod)

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = id
	claims["name"] = email
	claims["role_ids"] = roleIds
	claims["exp"] = expTime.Unix()

	t, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return "", err
	}

	return t, err
}

func NewRefreshTokenWithExp(id uuid.UUID, signingMethod jwt.SigningMethod, expTime time.Time) (string, error) {
	refreshToken := jwt.New(signingMethod)

	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["sub"] = id
	rtClaims["exp"] = expTime.Unix()

	rt, err := refreshToken.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return "", err
	}

	return rt, err
}

func NewAccessToken(id uuid.UUID, roleIds uuid.UUIDs, email string, signingMethod jwt.SigningMethod) (string, error) {
	// Create token
	var expireMinutes int64
	token := jwt.New(signingMethod)
	envExp := os.Getenv("EXPIRATION_MINUTES")
	expireMinutesInt, err := type_helper.ParseIntegerFromString[int64](envExp)
	if err != nil {
		slog.Error("Error parsing JWT Expiration minutes from env var EXPIRATION_MINUTS")
		expireMinutes = int64(15)
	}
	expireMinutes = expireMinutesInt

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = id
	claims["name"] = email
	claims["role_ids"] = roleIds
	claims["exp"] = time.Now().Add(time.Minute * time.Duration(expireMinutes)).Unix()

	t, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return "", err
	}

	return t, err
}

func NewRefreshToken(id uuid.UUID, signingMethod jwt.SigningMethod) (string, error) {
	refreshToken := jwt.New(signingMethod)
	refrshLengthEnv := os.Getenv("REFRESH_TOKEN_EXPIRIRATION_MINUTES")
	expireMinutesInt, err := type_helper.ParseIntegerFromString[int64](refrshLengthEnv)
	if err != nil {
		slog.Error("Error parsing JWT Expiration minutes from env var EXPIRATION_MINUTS")
		expireMinutesInt = int64(2880)
	}

	rtClaims := refreshToken.Claims.(jwt.MapClaims)
	rtClaims["sub"] = id
	rtClaims["exp"] = time.Now().Add(time.Hour * time.Duration(expireMinutesInt)).Unix()

	rt, err := refreshToken.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return "", err
	}

	return rt, err
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

func (a *LocalAuthService) GetUserById(id uuid.UUID) (*user_crud_svc.UserDao, error) {
	qry := infra_db_pg.New(a.DbConn)
	usrInfo := &user_crud_svc.UserDao{Id: id}

	user, err := qry.GetUserById(context.Background(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no user found with id: %s", id.String())
		}
		slog.Error("Error geting user from db", slog.String("error", err.Error()))
		return usrInfo, fmt.Errorf("error retrieving user info from db id: %s error: %w", id.String(), err)
	}

	usrInfo.ParseUserWithRoleFromDb(user)

	if !user.Enabled || user.IsDeleted {
		return usrInfo, fmt.Errorf("id: %s username: %s is disabled or deleted", user.ID, user.Username.String)
	}

	return usrInfo, nil
}

func (a *LocalAuthService) RefreshAccessToken(refreshToken string) (AuthToken, error) {
	var tokenPair AuthToken
	// Parse the refresh token and validate
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_KEY")), nil
	})

	if err != nil {
		return tokenPair, fmt.Errorf("refresh token validation error %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uid, err := uuid.Parse(claims["sub"].(string))
		if err != nil {
			return tokenPair, fmt.Errorf("error parsing uuid from string id: %s error: %w", claims["sub"].(string), err)
		}

		tokenPair.UserID = uid

		usrInfo, err := a.GetUserById(uid)
		// Get the user record from database or
		// run through your business logic to verify if the user can log in
		if err == nil {
			signingMethod := getJwtSigningMenthodFromEnv()
			tokenPair.Token, err = NewAccessToken(usrInfo.Id, usrInfo.RoleIds, usrInfo.Email, signingMethod)
			if err != nil {
				return tokenPair, fmt.Errorf("error creating NewAccessToken %w", err)
			}
			tokenPair.Email = usrInfo.Email
			tokenPair.Username = usrInfo.UserName
			return tokenPair, nil

		} else {
			return tokenPair, err
		}
	}
	return tokenPair, err
}
