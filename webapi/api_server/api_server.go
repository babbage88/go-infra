package api_server

import (
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/services"
	customlogger "github.com/babbage88/go-infra/utils/logger"
	authapi "github.com/babbage88/go-infra/webapi/authapi"
	userapi "github.com/babbage88/go-infra/webapi/user_api_handlers"
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

func StartWebApiServer(authService *authapi.UserAuthService, userCRUDService *services.UserCRUDService, srvadr *string) error {
	envars := authService.Envars
	mux := http.NewServeMux()
	mux.HandleFunc("/requestcert", authapi.AuthMiddleware(envars, authapi.Renewcert_renew(envars)))
	mux.HandleFunc("/login", authapi.LoginHandler(authService))
	mux.HandleFunc("/create/user", authapi.AuthMiddleware(envars, userapi.CreateUser(userCRUDService)))
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
