package cert_renew

import "time"

// swagger:response CertificateDataRenewResponse
type CertificateDataRenewResponse struct {
	Body CertificateData
}

// swagger:model CertificateData
type CertificateData struct {
	DomainNames   []string `json:"domainName"`
	CertPEM       string   `json:"cert_pem"`
	ChainPEM      string   `json:"chain_pem"`
	Fullchain     string   `json:"fullchain_pem"`
	PrivKey       string   `json:"priv_key"`
	ZipDir        string   `json:"zipDir"`
	S3DownloadUrl string   `json:"s3DownloadUrl"`
}

// Login Request takes in the certificate renewal payload.
// swagger:parameters Renew
type CertDnsRenewReqWrapper struct {
	// in:body
	Body CertDnsRenewReq `json:"body"`
}

// swagger:model CertDnsRenewReq
type CertDnsRenewReq struct {
	DomainNames          []string      `json:"domainName"`
	AcmeEmail            string        `json:"acmeEmail"`
	AcmeUrl              string        `json:"acmeUrl"`
	ZipDir               string        `json:"zipDir"`
	PushS3               bool          `json:"pushS3"`
	Token                string        `json:"token"`
	RecursiveNameServers []string      `json:"recurseServers"`
	Timeout              time.Duration `json:"timeout"`
}
