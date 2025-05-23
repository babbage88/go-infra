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

type LoginHandler struct {
	Service AuthService `json:"authService"`
}

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
			jsonResponse, _ := json.Marshal(token)
			w.WriteHeader(http.StatusOK)
			w.Write(jsonResponse)
			return
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Invalid credentials", LoginResult.Result.Error)
		}
	}
}

// swagger:route POST /login Authentication LocalLogin
// Login a user and return token.
// responses:
//   200: AuthToken

func (l *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	LoginHandleFunc(l.Service)
}

func parseAuthHeader(w http.ResponseWriter, r *http.Request) (*TokenRefreshReq, error) {
	retVal := &TokenRefreshReq{}
	authHeader := r.Header.Get("Authorization")
	parts := strings.Split(authHeader, "Bearer ")
	if len(parts) != 2 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Malformed Token"))
		return retVal, fmt.Errorf("malformed token")
	}
	retVal.RefreshToken = parts[1]
	return retVal, nil
}

// swagger:route POST /token/refresh Authentication RefreshAccessToken
// Refresh accessTokens andreturn to client.
// responses:
//
//	200: AuthToken
func RefreshAuthTokens(ua AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseAuthHeader(w, r)
		if err != nil {
			slog.Error("Error parsing Bearer token from Authorization Header", slog.String("Error", err.Error()))
		}

		newtokens, err := ua.RefreshAccessToken(req.RefreshToken)
		if err != nil {
			slog.Error("Error refreshing auth tokens", slog.String("Error", err.Error()))
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized, please login."))
		}
		jsonResponse, _ := json.Marshal(newtokens)
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}
