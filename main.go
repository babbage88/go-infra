package main

import (
	"flag"
	"io"
	"log/slog"
	"net/http"

	infra_db "github.com/babbage88/go-infra/database/infra_db"

	docker_helper "github.com/babbage88/go-infra/utils/docker_helper"
	customlogger "github.com/babbage88/go-infra/utils/logger"
	webapi "github.com/babbage88/go-infra/webapi"
)

func main() {

	db_pw := docker_helper.GetSecret("DB_PW")
	le_ini := docker_helper.GetSecret("trahan.dev_token")

	if le_ini == "" {
		slog.Warn("Le auth blank")
	}

	dbConn := infra_db.NewDatabaseConnection(infra_db.WithDbHost("10.0.0.92"), infra_db.WithDbPassword(db_pw))

	db, err := infra_db.InitializeDbConnection(dbConn)
	if err != nil {
		slog.Error("Error Connecting to Database", slog.String("Error", err.Error()))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/getalldns", webapi.CreateDnsHttpHandlerWrapper(db))
	mux.HandleFunc("/requestcert", webapi.WithAuth(webapi.RenewCertHandler))
	mux.HandleFunc("/healthCheck", webapi.HealthCheckHandler)

	config := customlogger.NewCustomLogger()
	clog := customlogger.SetupLogger(config)

	srvport := flag.String("srvadr", ":8993", "Address and port that http server will listed on. :8993 is default")
	flag.Parse()

	clog.Info("Starting http server.")
	err = http.ListenAndServe(*srvport, mux)
	if err != nil {
		clog.Error("Failed to start server", slog.String("Error", err.Error()))
	}

	defer func() {
		if err := infra_db.CloseDbConnection(); err != nil {
			clog.Error("Failed to close the database connection: %v", err)
		}
	}()

	defer func() {
		if file, ok := clog.Handler().(io.Closer); ok {
			file.Close()
		}
	}()
}
