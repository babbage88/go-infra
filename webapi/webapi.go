package webapi

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	cloudflaredns "github.com/babbage88/go-infra/cloud_providers/cloudflare"
	infra_db "github.com/babbage88/go-infra/database"
)

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
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
