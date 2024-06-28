package webapi

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	cloudflaredns "github.com/babbage88/go-infra/cloud_providers/cloudflare"
	infra_db "github.com/babbage88/go-infra/database"
	"github.com/babbage88/go-infra/webutils/certhandler"
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
