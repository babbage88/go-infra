package webapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/babbage88/go-infra/auth/hashing"
	jwt_auth "github.com/babbage88/go-infra/auth/tokens"
	cloudflaredns "github.com/babbage88/go-infra/cloud_providers/cloudflare"
	infra_db "github.com/babbage88/go-infra/database/infra_db"
	db_models "github.com/babbage88/go-infra/database/models"
	"github.com/babbage88/go-infra/utils/env_helper"
	"github.com/babbage88/go-infra/webutils/certhandler"
	"github.com/golang-jwt/jwt/v5"
)

type CfCertRequestResponse struct {
	DomainName       string              `json:"domainName"`
	CertbotCmdOutput ParsedCertbotOutput `json:"certbotOutput"`
}

type ParsedCertbotOutput struct {
	CertificateInfo string `json:"certificateInfo"`
	Warnings        string `json:"warnings"`
	DebugLog        string `json:"debugLog"`
}

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
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
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

func CreateDnsHttpHandlerWrapper(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			enableCors(&w)
			return
		}

		enableCors(&w)
		var records []cloudflaredns.DnsRecordReq
		slog.Info("Sending DB Request", slog.String("Query", "test"))
		records, _ = infra_db.GetAllDnsRecords(db)

		// Serialize response to JSON
		jsonResponse, err := json.Marshal(records)
		if err != nil {
			http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
			return
		}

		// Set response headers and write JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
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

func RenewCertHandler(w http.ResponseWriter, r *http.Request) {
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

	slog.Info("Received POST request")

	var req certhandler.CertDnsRenewReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	slog.Info("Decoded request body", slog.String("DomainName", req.DomainName))

	// Call the Renew method
	cmdOutput := req.Renew()

	// Parse the output
	parsedOutput := parseCertbotOutput(cmdOutput)

	slog.Info("Renewal command executed", slog.String("Output", strings.Join(cmdOutput, "\n")))

	// Prepare the response
	resp := CfCertRequestResponse{
		DomainName:       req.DomainName,
		CertbotCmdOutput: parsedOutput,
	}

	slog.Info("Marshaling JSON response", slog.String("DomainName", resp.DomainName))
	// Serialize response to JSON
	jsonResponse, err := json.Marshal(resp)
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

func LoginHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			enableCors(&w)
			return
		}

		enableCors(&w)

		w.Header().Set("Content-Type", "application/json")

		var u db_models.User
		json.NewDecoder(r.Body).Decode(&u)
		fmt.Printf("The user request value %v", u)

		dbuser, err := infra_db.GetUserByUsername(db, u.Username)
		if err != nil {
			slog.Error("Error getting user from database", slog.String("Error", err.Error()))
		}
		verify_pw := hashing.VerifyPassword(u.Password, dbuser.Password)

		if verify_pw {
			token, err := jwt_auth.CreateTokenanAddToDb(db, u.Id)
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
			fmt.Fprint(w, "Invalid credentials")
		}
	}
}

func AuthMidleware(next http.HandlerFunc) http.HandlerFunc {
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
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}
				SECRETKEY := env_helper.NewDotEnvSource(env_helper.WithVarName("JWT_KEY")).GetEnvVarValue()
				return []byte(SECRETKEY), nil
			})

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				ctx := context.WithValue(r.Context(), "props", claims)
				// Access context values in handlers like this
				// props, _ := r.Context().Value("props").(jwt.MapClaims)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				slog.Error("Error validating tokem", slog.String("Error", err.Error()))
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
			}
		}

		slog.Info("Token has breen verified.", slog.String("Host", r.URL.Host), slog.String("Path", r.URL.Path))
	}
}
