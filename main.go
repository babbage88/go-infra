package main

import (
	"flag"
	"log/slog"

	infra_db "github.com/babbage88/go-infra/database/infra_db"
	api_aerver "github.com/babbage88/go-infra/webapi/api_server"

	docker_helper "github.com/babbage88/go-infra/utils/docker_helper"
)

func main() {

	srvport := flag.String("srvadr", ":8993", "Address and port that http server will listed on. :8993 is default")
	flag.Parse()

	db_pw := docker_helper.GetSecret("DB_PW")
	le_ini := docker_helper.GetSecret("trahan.dev_token")

	if le_ini == "" {
		slog.Warn("Le auth blank")
	}

	dbConn := infra_db.NewDatabaseConnection(infra_db.WithDbHost("10.0.0.92"), infra_db.WithDbPassword(db_pw))

	db, _ := infra_db.InitializeDbConnection(dbConn)

	api_aerver.StartWebApiServer(db, srvport)

	defer func() {
		if err := infra_db.CloseDbConnection(); err != nil {
			slog.Error("Failed to close the database connection: %v", err)
		}
	}()

}
