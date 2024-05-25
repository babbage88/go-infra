package main

import (
	"io"
	"os"

	cloudflaredns "github.com/babbage88/go-infra/cloud_providers/cloudflare"
	infra_db "github.com/babbage88/go-infra/database"
	customlogger "github.com/babbage88/go-infra/utils/logger"
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
		DnsRecords:    []cloudflaredns.DnsRecordReq{},
	}

	// Fetch the DNS records
	dns_records, err := cloudflaredns.GetCurrentRecords(czone)

	czone.DnsRecords = dns_records

	// infra_db.GetDnsRecordByName(db, "_acme-challenge.api.trahan.dev", "TXT")
	infra_db.InsertDnsRecords(db, *czone)

	defer func() {
		if err := infra_db.CloseDbConnection(); err != nil {
			clog.Error("Failed to close the database connection: %v", err)
		}
	}()
}
