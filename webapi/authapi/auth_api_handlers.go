package authapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/babbage88/go-infra/webutils/cert_renew"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// swagger:route POST /renew renew idOfrenewEndpoint
// Request/Renew ssl certificate via cloudflare letsencrypt. Uses DNS Challenge
// responses:
//   200: CertificateData
// produces:
// - application/json
// - application/zip

func Renewcert_renew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Received POST request for Cert Renewal")
		var req cert_renew.CertDnsRenewReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		slog.Info("Decoded request body", slog.String("DomainName", req.DomainNames[0]))

		// Pass envars to the Renew method
		req.Timeout = req.Timeout * time.Second
		cert_info, err := req.Renew()
		if err != nil {
			slog.Error("error renewing cert", slog.String("error", err.Error()))
		}

		slog.Info("Renewal command executed")

		// Prepare the response
		slog.Info("Marshaling JSON response", slog.String("DomainName", cert_info.DomainNames[0]))
		// Serialize response to JSON
		jsonResponse, err := json.Marshal(cert_info)
		if err != nil {
			slog.Error("Failed to marshal JSON response", slog.String("Error", err.Error()))
			http.Error(w, "Failed to marshal JSON response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Set response headers and write JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
		slog.Info("Response sent successfully")
	}
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

// swagger:route POST /login login idOfloginEndpoint
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

// swagger:route POST /token/refresh tokenRefresh idOftokenRefreshEndpoint
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
