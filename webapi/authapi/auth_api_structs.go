package authapi

import (
	"time"

	"github.com/babbage88/go-infra/services"
	"github.com/golang-jwt/jwt/v5"
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
	UserName string `json:"userName"`
	Password string `json:"password"`
	IsHashed bool   `json:"isHashed"`
}

type UserLoginResponse struct {
	Result   LoginResult      `json:"result"`
	UserInfo services.UserDao `json:"UserDao"`
}

type InfraJWTClaim struct {
	*jwt.RegisteredClaims
	UserInfo interface{}
}

type AuthToken struct {
	Id           int32     `json:"id"`
	UserID       int32     `json:"user_id"`
	Token        string    `json:"token"`
	Expiration   time.Time `json:"expiration"`
	CreatedAt    time.Time `json:"created_at"`
	LastModified time.Time `json:"last_modified"`
}
