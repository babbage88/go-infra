package main

import (
	"io"
	"log/slog"

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

	// Create a CloudflareDnsZone object with hardcoded values
	czone := &cloudflaredns.CloudflareDnsZone{
		BaseUrl:  "https://api.cloudflare.com/client/v4/zones/",
		ZoneId:   "your_zone_id",
		RecordId: "your_record_id_here",
		CfToken:  "your_token_here",
	}

	// Call the GetCurrentRecords function
	cloudflaredns.GetCurrentRecords(czone)
}
