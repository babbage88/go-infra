package cloudflaredns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type CloudflareDnsZone struct {
	BaseUrl       string        `json:"baseUrl"`
	ZoneId        string        `json:"zoneId"`
	CfToken       string        `json:"cfToken"`
	RecordRequest *DnsRecordReq `json:"recordRequest"`
	SecretsPath   string        `json:"secretsPath"`
}

type DnsRecordReq struct {
	Content     string `json:"content"`
	Name        string `json:"name"`
	Proxied     bool   `json:"proxied"`
	Type        string `json:"type"`
	Comment     string `json:"comment"`
	Ttl         int16  `json:"ttl"`
	DnsRecordId string `json:"dnsRecordId"`
}

func buildRequestUrl(czone *CloudflareDnsZone) string {
	var bUrl bytes.Buffer

	bUrl.WriteString(czone.BaseUrl)
	bUrl.WriteString(czone.ZoneId)
	bUrl.WriteString("/dns_records")
	if czone.RecordRequest.DnsRecordId != "" {
		bUrl.WriteString("/" + czone.RecordRequest.DnsRecordId)
	}

	url := bUrl.String()

	return url
}

func createPayload(record *DnsRecordReq) (io.Reader, error) {
	data, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	return strings.NewReader(string(data)), nil
}

func GetCurrentRecords(czone *CloudflareDnsZone) *http.Response {
	//url := "https://api.cloudflare.com/client/v4/zones/[dns_zone_id]/dns_records"

	url := buildRequestUrl(czone)

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
	//fmt.Println(string(body))

	return res

}

func GetDnsRecordDetails(czone *CloudflareDnsZone) *http.Response {
	//url := "https://api.cloudflare.com/client/v4/zones/zone_id/dns_records/dns_record_id"

	url := buildRequestUrl(czone)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+czone.CfToken)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()

	return res
}

func CreateDnsRecord(czone *CloudflareDnsZone) *http.Response {
	url := buildRequestUrl(czone)

	payload, err := createPayload(czone.RecordRequest)
	if err != nil {
		fmt.Println("Error creating payload:", err)
	}
	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+czone.CfToken)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()

	return res
}

func UpdateDnsRecord(czone *CloudflareDnsZone) *http.Response {
	url := buildRequestUrl(czone)

	payload, err := createPayload(czone.RecordRequest)
	if err != nil {
		fmt.Println("Error creating payload:", err)
	}
	req, _ := http.NewRequest("PATCH", url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+czone.CfToken)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()

	return res
}
