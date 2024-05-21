package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	cloudflaredns "git.trahan.dev/go-infra/cloud_providers/cloudflare"
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

	clog.Info("Starting Infra tasks", slog.String("Action", "Retrieving DNS Records"))
	var api_key string = os.Getenv("CLOUFLARE_DNS_KEY")
	var cf_zone_ID string = os.Getenv("CF_ZONE_ID")
	fmt.Println("API Auth Token: " + api_key)
	fmt.Println("Zone ID: " + cf_zone_ID)

	dnsreq := &cloudflaredns.DnsRecordReq{
		Content:     "10.0.0.32",
		Name:        "testgo",
		Proxied:     false,
		Type:        "A",
		Comment:     "Testing Golang",
		Ttl:         3600,
		DnsRecordId: "your_record_id_here",
	}

	// Create a CloudflareDnsZone object with hardcoded values
	czone := &cloudflaredns.CloudflareDnsZone{
		BaseUrl:       "https://api.cloudflare.com/client/v4/zones/",
		ZoneId:        cf_zone_ID,
		CfToken:       api_key,
		RecordRequest: dnsreq,
	}
	fmt.Println(fmt.Sprint(czone))
	// Call the GetCurrentRecords function
	//cloudflaredns.CreateDnsRecord(czone)
}
