package main

import (
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"

	cloudflaredns "git.trahan.dev/go-infra/cloud_providers/cloudflare"
	infra_db "git.trahan.dev/go-infra/database"
	customlogger "git.trahan.dev/go-infra/utils"
)

func main() {

	config := customlogger.NewCustomLogger()

	clog := customlogger.SetupLogger(config)

	defer func() {
		if file, ok := clog.Handler().(io.Closer); ok {
			file.Close()
		}
	}()

	var api_key string = os.Getenv("CLOUFLARE_DNS_KEY")
	var cf_zone_ID string = os.Getenv("CF_ZONE_ID")

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
	}

	// Fetch the DNS records
	dns_records, err := cloudflaredns.GetCurrentRecords(czone)

	// Database connection details
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Connect to the PostgreSQL database
	clog.Info("Connecting to database: " + dbHost)
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		clog.Error("Error connecting to the database", slog.String("Error", err.Error()))
		return
	}
	defer db.Close()

	if err != nil {
		clog.Error("Error retrieving DNS records", slog.String("Error", err.Error()))
		return
	}

	clog.Info("Inserting Records into Database: ")
	infra_db.InsertDnsRecords(db, dns_records)

}
