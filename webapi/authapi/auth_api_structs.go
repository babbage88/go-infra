package authapi

import (
	"time"

	"github.com/babbage88/go-infra/services/user_crud_svc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

// Login Request takes  in Username and Password.
// swagger:parameters LocalLogin
type UserLoginReqWrapper struct {
	// in:body
	Body UserLoginRequest `json:"body"`
}

type UserLoginRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type UserLoginResponse struct {
	Result   LoginResult           `json:"result"`
	UserInfo user_crud_svc.UserDao `json:"UserDao"`
}

type InfraJWTClaim struct {
	*jwt.RegisteredClaims
	UserInfo interface{}
}

// Respose will return login result and the user info.
// swagger:response AuthToken
// This text will appear as description of your response body.
type AuthToken struct {
	// in:body
	UserID       uuid.UUID `json:"user_id"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	Expiration   time.Time `json:"expiration"`
}

type TokenRefreshReq struct {
	RefreshToken string `json:"refreshToken"`
}

type UserPermission struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
