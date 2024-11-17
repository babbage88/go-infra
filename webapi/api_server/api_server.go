package api_server

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/utils/env_helper"
	customlogger "github.com/babbage88/go-infra/utils/logger"
	authapi "github.com/babbage88/go-infra/webapi/authapi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Package classification goinfra.
//
// Go-Infra API forInfrastruction Automation.
//
//     Schemes: http
//     BasePath: /
//     Version: 1.0.5
//     Host: infra.test.trahan.dev
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Security:
//     - basic
//
//    SecurityDefinitions:
//    basic:
//      type: basic
//
// swagger:meta

func StartWebApiServer(envars *env_helper.EnvVars, db *sql.DB, srvadr *string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/getalldns", authapi.AuthMiddleware(envars, authapi.CreateDnsHttpHandlerWrapper(db)))
	mux.HandleFunc("/requestcert", authapi.AuthMiddleware(envars, authapi.Renewcert_renew(envars)))
	mux.HandleFunc("/login", authapi.LoginHandler(envars, db))
	mux.HandleFunc("/healthCheck", authapi.HealthCheckHandler)
	mux.Handle("/metrics", promhttp.Handler())

	config := customlogger.NewCustomLogger()
	clog := customlogger.SetupLogger(config)

	clog.Info("Starting http server.")
	err := http.ListenAndServe(*srvadr, mux)
	if err != nil {
		slog.Error("Failed to start server", slog.String("Error", err.Error()))
	}
	return err
}
