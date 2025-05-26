package cert_renew

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// swagger:route POST /renew Certificates Renew
// Request/Renew ssl certificate via cloudflare letsencrypt. Uses DNS Challenge
// responses:
//  200: CertificateDataRenewResponse
//	400: description:Bad Request
//	401: description:Unauthorized
//	500: description:Insernal Server Error
// produces:
// - application/json
// - application/zip

func Renewcert_renew() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Received POST request for Cert Renewal")
		var req CertDnsRenewReq
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
