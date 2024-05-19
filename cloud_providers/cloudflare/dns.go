package cloudflaredns

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type CloudflareDnsZone struct {
	BaseUrl  string
	ZoneId   string
	RecordId string
	CfToken  string
}

func GetCurrentRecords(czone *CloudflareDnsZone) {
	//url := "https://api.cloudflare.com/client/v4/zones/[dns_zone_id]/dns_records"
	var bUrl bytes.Buffer

	bUrl.WriteString(czone.BaseUrl)
	bUrl.WriteString(czone.ZoneId)
	bUrl.WriteString("/dns_records")

	url := bUrl.String()

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+czone.CfToken)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	// fmt.Println(res)
	slog.Debug(fmt.Sprint(res))
	slog.Debug(fmt.Sprint(string(body)))
	slog.Info("Retrieving Current DNS Records")

}
