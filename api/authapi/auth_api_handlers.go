package authapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

type LoginHandler struct {
	Service AuthService `json:"authService"`
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

// swagger:route POST /token/refresh Authentication RefreshAccessToken
// Refresh accessTokens and return to client.
// responses:
//
//	200: RefreshAccessTokenResponse
//	400: description:Bad Request
//	401: description:Unauthorized
//	500: description:Insernal Server Error
func RefreshAccessTokens(ua AuthService) http.HandlerFunc {
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
		resp := AccessTokenRefreshResponse{AccessToken: newtokens.Token, RefreshToken: refreshReq.RefreshToken}
		w.Header().Set("Content-Type", "application/json")
		jsonResponse, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "error marshaling response", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}
