package api_server

import (
	"database/sql"
	"log/slog"
	"net/http"

	customlogger "github.com/babbage88/go-infra/utils/logger"
	webapi "github.com/babbage88/go-infra/webapi/api_handlers"
)

func StartWebApiServer(db *sql.DB, srvadr *string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/getalldns", webapi.AuthMidleware(webapi.CreateDnsHttpHandlerWrapper(db)))
	mux.HandleFunc("/requestcert", webapi.AuthMidleware(webapi.RenewCertHandler))
	mux.HandleFunc("/login", webapi.LoginHandler(db))
	mux.HandleFunc("/healthCheck", webapi.HealthCheckHandler)

	config := customlogger.NewCustomLogger()
	clog := customlogger.SetupLogger(config)

	clog.Info("Starting http server.")
	err := http.ListenAndServe(*srvadr, mux)
	if err != nil {
		slog.Error("Failed to start server", slog.String("Error", err.Error()))
	}
	return err
}
