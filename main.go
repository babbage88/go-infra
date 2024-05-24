package main

import (
	"io"
	"os"

	cloudflaredns "git.trahan.dev/go-infra/cloud_providers/cloudflare"
	infra_db "git.trahan.dev/go-infra/database"
	customlogger "git.trahan.dev/go-infra/utils/logger"
)

func main() {

	config := customlogger.NewCustomLogger()

	clog := customlogger.SetupLogger(config)

	defer func() {
		if file, ok := clog.Handler().(io.Closer); ok {
			file.Close()
		}
	}()

	dbConn := infra_db.NewDatabaseConnection(infra_db.WithDbHost("10.0.0.92"), infra_db.WithDbPassword(os.Getenv("DB_PASSWORD")))

	db, err := infra_db.InitializeDbConnection(dbConn)

	if err != nil {
		clog.Error("Error Connecting to Database", err)
	}

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

	clog.Info("Inserting Records into Database: ")
	infra_db.GetDnsRecordByName(db, "_acme-challenge.api.trahan.dev", "TXT")
	infra_db.InsertDnsRecords(db, dns_records)

	defer func() {
		if err := infra_db.CloseDbConnection(); err != nil {
			clog.Error("Failed to close the database connection: %v", err)
		}
	}()

	/*
		// Example of using GetDnsRecordByName
		name := "trahan.dev"
		record, err := infra_db.GetDnsRecordByName(db, name, dnsreq.Type)
		if err != nil {
			clog.Error("Error fetching DNS record by name", slog.String("Error", err.Error()))
			return
		}

		if record != nil {
			clog.Info("DNS Record found:", slog.String("name", record.Name), slog.String("content", record.Content), slog.String("Id", record.DnsRecordId))
		} else {
			clog.Info("No DNS record found with the specified name", slog.String("name", name))
		}

		//cloudflaredns.GetDnsRecordDetails(czone)
	*/
}
