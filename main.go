package main

import (
	"flag"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"

	cloudflaredns "github.com/babbage88/go-infra/cloud_providers/cloudflare"
	infra_db "github.com/babbage88/go-infra/database"
	_ "github.com/babbage88/go-infra/swagger"
	docker_helper "github.com/babbage88/go-infra/utils/docker_helper"
	customlogger "github.com/babbage88/go-infra/utils/logger"
	webapi "github.com/babbage88/go-infra/webapi"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title go-infra api
// @version 1.0
// @description Automation for hybrid web app deployments

// @contact.name Justin
// @contact.url trahan@babbage88
// @contact.email support@swagger.io

// @host localhost:8993
// @BasePath /
// @query.collection.format multi
func main() {
	db_pw := docker_helper.GetSecret("DB_PW")
	api_key := docker_helper.GetSecret("trahan.dev_token")
	cf_zone_ID := docker_helper.GetSecret("trahan.dev_zoneid")

	dbConn := infra_db.NewDatabaseConnection(infra_db.WithDbHost("10.0.0.92"), infra_db.WithDbPassword(db_pw))

	db, err := infra_db.InitializeDbConnection(dbConn)

	if err != nil {
		slog.Error("Error Connecting to Database", slog.String("Error", err.Error()))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)
	// @Summary Return http Handler function to return all DNS records from Database
	// @ID GetAlldns
	// @Produce  json
	// @Success 200 {object} []cloudflaredns.DnsRecordReq
	// @Router /api/getalldns [get]
	mux.HandleFunc("/getalldns/", webapi.CreateDnsHttpHandlerWrapper(db))

	config := customlogger.NewCustomLogger()
	clog := customlogger.SetupLogger(config)

	srvport := flag.String("srvadr", ":8993", "Address and port that http server will listed on. :8993 is default")
	flag.Parse()

	clog.Info("Starting http server.")
	http.ListenAndServe(*srvport, mux)

	app := &cli.App{
		Name:  "boom",
		Usage: "make an explosive entrance",
		Action: func(*cli.Context) error {

			dnsreq := &cloudflaredns.DnsRecordReq{
				Content:     "10.0.0.32",
				Name:        "testgo",
				Proxied:     false,
				Type:        "A",
				Comment:     "Testing Golang",
				Ttl:         3600,
				DnsRecordId: "",
			}

			// Create a CloudflareDnsZone object with hardcoded values
			czone := &cloudflaredns.CloudflareDnsZone{
				BaseUrl:       "https://api.cloudflare.com/client/v4/zones/",
				ZoneId:        cf_zone_ID,
				CfToken:       api_key,
				RecordRequest: dnsreq,
				DnsRecords:    []cloudflaredns.DnsRecordReq{},
			}

			dns_records, err := cloudflaredns.GetCurrentRecords(czone)
			if err != nil {
				slog.Error("Error getting DNS rocords from CF", slog.String("Error", err.Error()))
			}

			czone.DnsRecords = dns_records

			infra_db.InsertDnsRecords(db, *czone)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("Error Running command", slog.String("Error", err.Error()))
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
