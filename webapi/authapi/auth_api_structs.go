package authapi

import (
	"github.com/babbage88/go-infra/database/services"
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
