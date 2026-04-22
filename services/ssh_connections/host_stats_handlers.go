package ssh_connections

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/api/authapi"
	"github.com/google/uuid"
)

func (m *SSHConnectionManager) HostStatsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := authapi.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stats, err := m.CollectAllHostStats(r.Context(), userID)
	if err != nil {
		slog.Error("failed to collect host stats", slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func (m *SSHConnectionManager) HostStatsByIDHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := authapi.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	hostServerID, err := uuid.Parse(r.PathValue("ID"))
	if err != nil {
		http.Error(w, "Invalid host server ID", http.StatusBadRequest)
		return
	}

	stats := m.CollectHostStats(r.Context(), userID, hostServerID)
	status := http.StatusOK
	if stats.Status == "error" {
		status = http.StatusBadGateway
	}

	writeJSON(w, status, stats)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to encode response", slog.String("error", err.Error()))
	}
}
