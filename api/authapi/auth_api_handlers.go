package authapi

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// swagger:route POST /login Authentication LocalLogin
// Local Auth login with username and password.
// Successful login sets secure HTTP-only auth cookies for subsequent requests.
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
				return
			}
			setAuthCookies(w, token.Token, token.RefreshToken, token.Expiration)
			response := LocalLoginResponse{UserID: LoginResult.UserInfo.Id,
				Username: LoginResult.UserInfo.UserName, Email: LoginResult.UserInfo.Email,
				Expiration: token.Expiration}
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
// When a refresh cookie is present, the request body token is optional.
// responses:
//
//	200: RefreshAccessTokenResponse
//	400: description:Bad Request
//	401: description:Unauthorized
//	500: description:Insernal Server Error
func RefreshAccessTokensHandleFunc(ua AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var refreshReq TokenRefreshReq
		if r.Body != nil {
			err := json.NewDecoder(r.Body).Decode(&refreshReq)
			if err != nil && err != io.EOF {
				slog.Error("Error parsing refresh token from request body", slog.String("Error", err.Error()))
				http.Error(w, "error parsing refresh token from request body", http.StatusBadRequest)
				return
			}
		}

		if refreshReq.RefreshToken == "" {
			refreshReq.RefreshToken = GetRefreshTokenFromRequest(r)
		}
		if refreshReq.RefreshToken == "" {
			http.Error(w, "missing refresh token", http.StatusUnauthorized)
			return
		}

		newtokens, err := ua.RefreshAccessToken(refreshReq.RefreshToken)
		if err != nil {
			slog.Error("Error refreshing auth tokens", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized, please login."))
			return
		}
		setAuthCookies(w, newtokens.Token, refreshReq.RefreshToken, newtokens.Expiration)
		resp := AccessTokenRefreshResponse{AccessToken: newtokens.Token,
			UserID:   newtokens.UserID,
			Username: newtokens.Username,
			Email:    newtokens.Email,
		}
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "error marshaling response", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

func RefreshAccessTokensHandler(ua AuthService) http.Handler {
	return http.HandlerFunc(RefreshAccessTokensHandleFunc(ua))
}

// swagger:route POST /token/verify Authentication VerifyToken
// Verify the current access token's validity.
// The token may be supplied by the secure auth cookie or the Authorization header.
// responses:
//
//	200: description:Valid Token
//	401: description:Unauthorized
//	400: description:Bad Request

func VerifyTokenHandler(auth_svc AuthService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tokenString := GetAccessTokenFromRequest(r)
		if tokenString == "" {
			http.Error(w, `{"error":"Authentication token missing"}`, http.StatusUnauthorized)
			return
		}

		err := auth_svc.VerifyToken(tokenString)
		if err != nil {
			slog.Warn("Token verification failed", slog.String("error", err.Error()))
			http.Error(w, `{"error":"Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"status": "token valid"})
	})
}

func LogoutHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clearAuthCookies(w)
		w.WriteHeader(http.StatusNoContent)
	})
}

func SessionHandler(authSvc AuthService) http.Handler {
	return AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, `{"error":"Unable to determine user from token"}`, http.StatusUnauthorized)
			return
		}

		user, err := authSvc.GetUserById(userID)
		if err != nil {
			http.Error(w, `{"error":"Unable to load current user"}`, http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SessionInfoResponse{
			UserID:   user.Id,
			Username: user.UserName,
			Email:    user.Email,
			Roles:    user.Roles,
		})
	}))
}
