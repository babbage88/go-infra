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

// Login Request takes  in Username and Password.
// swagger:parameters idOfloginEndpoint
// in:body
type UserLoginRequest struct {
	// in:body
	UserName string `json:"userName"`
	// in:body
	Password string `json:"password"`
}

// swagger:route POST /login login-tag idOfloginEndpoint
// Login a user and return token.
// responses:
//   200: UserLoginResponse

// Respose will return login result and the user info.
// swagger:response UserLoginResponse
type UserLoginResponse struct {
	// This text will appear as description of your response body.
	// in:body
	Result LoginResult `json:"result"`
	// in:body
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
