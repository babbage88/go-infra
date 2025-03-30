package cert_renew

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/babbage88/go-acme-cli/cloud_providers/cf_acme"
)

var certPrefix string = "-----BEGIN CERTIFICATE-----\n"

var certSuffix string = "\n-----END CERTIFICATE-----\n"

var keyPrefix string = "-----BEGIN PRIVATE KEY-----\n"

var keySuffix string = "\n-----END PRIVATE KEY-----\n"

type CertificateData struct {
	DomainNames   []string `json:"domainName"`
	CertPEM       string   `json:"cert_pem"`
	ChainPEM      string   `json:"chain_pem"`
	Fullchain     string   `json:"fullchain_pem"`
	PrivKey       string   `json:"priv_key"`
	ZipDir        string   `json:"zipDir"`
	S3DownloadUrl string   `json:"s3DownloadUrl"`
}

// Login Request takes  in Username and Password.
// swagger:parameters idOfrenewEndpoint
type CertDnsRenewReqWrapper struct {
	// in:body
	Body CertDnsRenewReq `json:"body"`
}

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

func (c *CertDnsRenewReq) InitAcmeRenewRequest() *cf_acme.CertificateRenewalRequest {
	cfReq := &cf_acme.CertificateRenewalRequest{
		DomainNames:          c.DomainNames,
		AcmeEmail:            c.AcmeEmail,
		AcmeUrl:              c.AcmeUrl,
		PushS3:               c.PushS3,
		ZipDir:               c.ZipDir,
		Token:                c.Token,
		RecursiveNameServers: c.RecursiveNameServers,
		Timeout:              c.Timeout,
	}

	return cfReq
}

func (c *CertificateData) ParseAcmeCertStruct(acmeCert *cf_acme.CertificateData) {
	c.DomainNames = acmeCert.DomainNames
	c.CertPEM = acmeCert.CertPEM
	c.ChainPEM = acmeCert.CertPEM
	c.Fullchain = acmeCert.Fullchain
	c.PrivKey = acmeCert.PrivKey
	c.ZipDir = acmeCert.ZipDir
	c.S3DownloadUrl = acmeCert.S3DownloadUrl
}
func (c *CertificateData) TrimJsonCertificateData() {
	certTrimmed, err := readAndTrimCert(c.CertPEM, certPrefix, certSuffix)
	if err == nil {
		slog.Info("Timming certificate prefix/suffix json")
		c.CertPEM = certTrimmed
	}
	priKeyStr, err := readAndTrimCert(c.PrivKey, keyPrefix, keySuffix)
	if err == nil {
		slog.Info("Trimming key prefix for json")
		c.PrivKey = priKeyStr
	}
	chainStrTrim, err := readAndTrimCert(c.ChainPEM, certPrefix, certSuffix)
	if err == nil {
		slog.Info("Timming certificate prefix/suffix json")
		c.ChainPEM = chainStrTrim
	}
}

type Renewal interface {
	Renew(token string, recursiveNameservers []string, timeout time.Duration) cf_acme.CertificateData
}

// CertRenewReq is the interface with the GetDomainName method.
type CertRenewReq interface {
	GetDomainName() string
}

func (c *CertDnsRenewReq) Renew() (*CertificateData, error) {
	certData := &CertificateData{}
	acmeRenewal := c.InitAcmeRenewRequest()
	certificates, err := acmeRenewal.Renew(c.Token, c.RecursiveNameServers, c.Timeout)
	if err != nil {
		slog.Error("error renewing certificate")
	}
	err = certificates.PushCertBufferToS3(c.ZipDir)
	if err != nil {
		slog.Error("error pushing zip file to S3", slog.String("error", err.Error()))
	}
	err = os.Remove(certificates.ZipDir)
	if err != nil {
		slog.Error("error removing zip file", slog.String("error", err.Error()))
	}

	certData.ParseAcmeCertStruct(&certificates)

	return certData, err
}

// ReadAndTrimFile reads the content of a file, removes the PEM delimiters, and returns the trimmed content.
func readAndTrimCert(s string, beginMarker string, endMarker string) (string, error) {
	s = strings.ReplaceAll(s, beginMarker, "")
	s = strings.ReplaceAll(s, endMarker, "")
	slog.Info("Paring finished", slog.String("Content", s))

	return strings.TrimSpace(s), nil
}
