package authapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// swagger:route POST /login Authentication LocalLogin
// Local Auth login with username and password
// responses:
//
//	200: LocalLoginResponse
//	400: description:Bad Request
//	401: description:Unauthorized
//	500: description:Insernal Server Error

func LoginHandleFunc(auth_svc AuthService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Username and Pasword in request body as json.
		// in:body
		var loginReq *UserLoginRequest
		json.NewDecoder(r.Body).Decode(&loginReq)
		LoginResult := auth_svc.Login(loginReq)

		if LoginResult.Result.Success {
			token, err := auth_svc.CreateAuthTokenOnLogin(LoginResult.UserInfo.Id, LoginResult.UserInfo.RoleIds, LoginResult.UserInfo.Email)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				slog.Error("Error verifying password", slog.String("Error", err.Error()))
			}
			response := LocalLoginResponse{UserID: LoginResult.UserInfo.Id,
				Username: LoginResult.UserInfo.UserName, Email: LoginResult.UserInfo.Email,
				Token: token.Token, RefreshToken: token.RefreshToken, Expiration: token.Expiration}
			jsonResponse, _ := json.Marshal(response)
			w.WriteHeader(http.StatusOK)
			w.Write(jsonResponse)
			return
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Invalid credentials", LoginResult.Result.Error)
		}
	}
}

func LoginHandler(auth_svc AuthService) http.Handler {
	return http.HandlerFunc(LoginHandleFunc(auth_svc))
}

// swagger:route POST /token/refresh Authentication RefreshAccessToken
// Refresh accessTokens and return to client.
// responses:
//
//	200: RefreshAccessTokenResponse
//	400: description:Bad Request
//	401: description:Unauthorized
//	500: description:Insernal Server Error
func RefreshAccessTokensHandleFunc(ua AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var refreshReq TokenRefreshReq
		err := json.NewDecoder(r.Body).Decode(&refreshReq)
		if err != nil {
			slog.Error("Error parsing refresh token from request body", slog.String("Error", err.Error()))
			http.Error(w, "error parsing refresh token from request body", http.StatusBadRequest)
		}

		newtokens, err := ua.RefreshAccessToken(refreshReq.RefreshToken)
		if err != nil {
			slog.Error("Error refreshing auth tokens", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized, please login."))
		}
		resp := AccessTokenRefreshResponse{AccessToken: newtokens.Token,
			RefreshToken: refreshReq.RefreshToken,
			UserID:       newtokens.UserID,
			Username:     newtokens.Username,
			Email:        newtokens.Email,
		}
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "error marshaling response", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

func RefreshAccessTokensHandler(ua AuthService) http.Handler {
	return http.HandlerFunc(RefreshAccessTokensHandleFunc(ua))
}

// swagger:route POST /token/verify Authentication VerifyToken
// Verify a JWT access token's validity.
// responses:
//
//	200: description:Valid Token
//	401: description:Unauthorized
//	400: description:Bad Request

func VerifyTokenHandler(auth_svc AuthService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"Authorization header missing"}`, http.StatusBadRequest)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, `{"error":"Authorization header must be in 'Bearer <token>' format"}`, http.StatusBadRequest)
			return
		}

		tokenString := parts[1]

		err := auth_svc.VerifyToken(tokenString)
		if err != nil {
			slog.Warn("Token verification failed", slog.String("error", err.Error()))
			http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "token valid"})
	})
}
