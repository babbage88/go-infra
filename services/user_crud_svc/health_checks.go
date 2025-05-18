package user_crud_svc

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DbHeathCheckResponse struct {
	Error     error  `json:"error"`
	Status    string `json:"status"`
	CheckType string `json:"checkType"`
	Id        int32  `json:"id"`
}

type HealthCheckService struct {
	DbConn *pgxpool.Pool
}

type IHealthCheckService interface {
	DbReadHealthCheck() DbHeathCheckResponse
	DbReadHealthCheckHandler() func(http.ResponseWriter, *http.Request)
}

func (h *HealthCheckService) DbReadHealthCheck() DbHeathCheckResponse {
	dbHealth := &DbHeathCheckResponse{CheckType: "Read"}
	queries := infra_db_pg.New(h.DbConn)
	qry, err := queries.DbHealthCheckRead(context.Background())
	if err != nil {
		slog.Error("Error executing DbReadHealthCheck query", slog.String("error", err.Error()))
		dbHealth.Error = err
		return *dbHealth
	}

	dbHealth.Id = qry.ID
	dbHealth.Status = qry.Status.String
	dbHealth.Error = nil
	return *dbHealth
}

func (h *HealthCheckService) DbReadHealthCheckHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		readCheck := h.DbReadHealthCheck()
		if readCheck.Error != nil {
			slog.Error("Error running db read healthcheck", slog.String("error", readCheck.Error.Error()))
			http.Error(w, "Failed to run database healthcheck query: "+readCheck.Error.Error(), http.StatusInternalServerError)
			return
		}
		hcResponse, err := json.Marshal(readCheck)
		if err != nil {
			slog.Error("Error marshing response", slog.String("error", err.Error()))
			http.Error(w, "error marshaling response"+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(hcResponse)
		slog.Info("Response sent successfully")
	}
}
