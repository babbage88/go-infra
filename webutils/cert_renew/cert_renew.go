package cert_renew

import (
	"archive/zip"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/babbage88/go-acme-cli/cloud_providers/cf_acme"
)

var certPrefix string = "-----BEGIN CERTIFICATE-----"

var certSuffix string = "-----END CERTIFICATE-----"

var keyPrefix string = "-----BEGIN PRIVATE KEY-----"

var keySuffix string = "-----END PRIVATE KEY-----"

var certbotConfigDir string = ".certbot/config"

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
	SaveZip              bool          `json:"saveZip"`
	ZipDir               string        `json:"zipDir"`
	PushS3               bool          `json:"pushS3"`
	Token                string        `json:"token"`
	RecursiveNameServers []string      `json:"recurseServers"`
	Timeout              time.Duration `json:"timeout"`
}

func (c *CertDnsRenewReq) InitAcmeRenewRequest() *cf_acme.CertificateRenewalRequest {
	cfReq := &cf_acme.CertificateRenewalRequest{
		DomainNames: c.DomainNames,
		AcmeEmail:   c.AcmeEmail,
		AcmeUrl:     c.AcmeUrl,
		SaveZip:     c.SaveZip,
		PushS3:      c.PushS3,
		ZipDir:      c.ZipDir,
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
	c.S3DownloadUrl = c.S3DownloadUrl
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
	certificates, err := acmeRenewal.RenewCertWithDns()
	if err != nil {
		slog.Error("error renewing certificate")
	}
	certificates.PushZipDirToS3(c.ZipDir)
	certData.ParseAcmeCertStruct(&certificates)

	certData.TrimJsonCertificateData()
	return certData, err
}

// ReadAndTrimFile reads the content of a file, removes the PEM delimiters, and returns the trimmed content.
func readAndTrimCert(s string, beginMarker string, endMarker string) (string, error) {
	s = strings.ReplaceAll(s, beginMarker, "")
	s = strings.ReplaceAll(s, endMarker, "")
	slog.Info("Paring finished", slog.String("Content", s))

	return strings.TrimSpace(s), nil
}

func createZipFile(liveDir string, certInfo CertificateData) error {
	zipFileName := fmt.Sprintf("%s/certs.zip", liveDir)
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// List of files to include in the zip
	files := []struct {
		Name, Content string
	}{
		{"cert.pem", certInfo.CertPEM},
		{"chain.pem", certInfo.ChainPEM},
		{"fullchain.pem", certInfo.Fullchain},
		{"privkey.pem", certInfo.PrivKey},
	}

	for _, file := range files {
		f, err := zipWriter.Create(file.Name)
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(file.Content))
		if err != nil {
			return err
		}
	}

	return nil
}
