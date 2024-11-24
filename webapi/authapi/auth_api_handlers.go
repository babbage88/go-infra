package authapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/babbage88/go-infra/utils/env_helper"
	"github.com/babbage88/go-infra/webutils/cert_renew"
	"github.com/golang-jwt/jwt/v5"
)

const apiToken = "your-secret-api-token"

func WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "Bearer "+apiToken {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "Authorization ,origin, content-type, accept, x-requested-with")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		enableCors(&w)
		return
	}
	enableCors(&w)
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

func Renewcert_renew(envars *env_helper.EnvVars) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		if r.Method == "OPTIONS" {
			slog.Info("Received OPTIONS request")
			return
		}

		if r.Method != http.MethodPost {
			slog.Error("Invalid request method", slog.String("Method", r.Method))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		slog.Info("Received POST request for Cert Renewal")

		var req cert_renew.CertDnsRenewReq
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		slog.Info("Decoded request body", slog.String("DomainName", req.DomainName))

		// Pass envars to the Renew method
		cert_info, err := req.Renew(envars)

		slog.Info("Renewal command executed")

		// Prepare the response

		slog.Info("Marshaling JSON response", slog.String("DomainName", cert_info.DomainName))
		// Serialize response to JSON
		jsonResponse, err := json.Marshal(cert_info)
		if err != nil {
			slog.Error("Failed to marshal JSON response", slog.String("Error", err.Error()))
			http.Error(w, "Failed to marshal JSON response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if req.ZipFiles {
			w.Header().Set("Content-Type", "application/zip")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", req.DomainName))
			http.ServeFile(w, r, cert_info.ZipDir)
		} else {

			// Set response headers and write JSON response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(jsonResponse)
			slog.Info("Response sent successfully")
		}
	}
}

// swagger:route POST /login login idOfloginEndpoint
// Login a user and return token.
// responses:
//   200: AuthToken

func LoginHandler(ua_service *UserAuthService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			enableCors(&w)
			return
		}

		enableCors(&w)

		w.Header().Set("Content-Type", "application/json")
		// Username and Pasword in request body as json.
		// in:body
		var loginReq *UserLoginRequest
		json.NewDecoder(r.Body).Decode(&loginReq)
		LoginResult := loginReq.Login(ua_service.DbConn)

		if LoginResult.Result.Success {
			token, err := ua_service.CreateAuthToken(LoginResult.UserInfo.Id, LoginResult.UserInfo.Role, LoginResult.UserInfo.Email)
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

func AuthMiddleware(envars *env_helper.EnvVars, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			enableCors(&w)
			return
		}

		enableCors(&w)

		w.Header().Set("Content-Type", "application/json")
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			fmt.Println("Malformed token")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Malformed Token"))
		} else {
			jwtToken := authHeader[1]
			token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				// Retrieve the secret key from environment variables
				SECRETKEY := envars.GetVarMapValue("JWT_KEY")
				if SECRETKEY == "" {
					return nil, fmt.Errorf("secret key not found")
				}
				return []byte(SECRETKEY), nil
			})

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				ctx := context.WithValue(r.Context(), "props", claims)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				slog.Error("Error validating token", slog.String("Error", err.Error()))
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
			}
		}

		slog.Info("Token has been verified.", slog.String("Host", r.URL.Host), slog.String("Path", r.URL.Path))
	}
}

func setCookieHandler(w http.ResponseWriter, r *http.Request, token string) {
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
