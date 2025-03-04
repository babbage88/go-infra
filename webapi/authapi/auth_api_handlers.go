package authapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/babbage88/go-infra/webutils/cert_renew"
	"github.com/golang-jwt/jwt/v5"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func parseCertbotOutput(output []string) ParsedCertbotOutput {
	var certInfo, warnings, debugLog string

	for _, line := range output {
		if strings.Contains(line, "Saving debug log") {
			debugLog += line + "\n"
		} else if strings.Contains(line, "Unsafe permissions on credentials configuration file") {
			warnings += line + "\n"
		} else {
			certInfo += line + "\n"
		}
	}

	return ParsedCertbotOutput{
		CertificateInfo: certInfo,
		Warnings:        warnings,
		DebugLog:        debugLog,
	}
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

// swagger:route POST /login login idOfloginEndpoint
// Login a user and return token.
// responses:
//   200: AuthToken

func LoginHandler(ua_service *UserAuthService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Username and Pasword in request body as json.
		// in:body
		var loginReq *UserLoginRequest
		json.NewDecoder(r.Body).Decode(&loginReq)
		LoginResult := loginReq.Login(ua_service.DbConn)

		if LoginResult.Result.Success {
			token, err := ua_service.CreateAuthTokenOnLogin(LoginResult.UserInfo.Id, LoginResult.UserInfo.RoleIds, LoginResult.UserInfo.Email)
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
func RefreshAuthTokens(ua *UserAuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jwtKey := os.Getenv("JWT_KEY")
		req, err := parseAuthHeader(w, r)
		if err != nil {
			slog.Error("Error parsing Bearer token from Authorization Header", slog.String("Error", err.Error()))
		}
		token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtKey), nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if sub, ok := claims["sub"].(float64); ok {
				uid := int32(sub)
				newtokens := &AuthToken{UserID: uid, RefreshToken: req.RefreshToken}
				err := newtokens.RefreshAccessTokens(ua.DbConn)
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
	}
}

func setCookieHandler(w http.ResponseWriter, token string) {
	// Initialize a new cookie containing the string "Hello world!" and some
	// non-default attributes.
	cookie := http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	// Use the http.SetCookie() function to send the cookie to the client.
	// Behind the scenes this adds a `Set-Cookie` header to the response
	// containing the necessary cookie data.
	http.SetCookie(w, &cookie)

	// Write a HTTP response as normal.
	w.Write([]byte("cookie set!"))
}

func getCookieHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			http.Error(w, "cookie not found", http.StatusBadRequest)
		default:
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}

	// Echo out the cookie value in the response body.
	w.Write([]byte(cookie.Value))
}
