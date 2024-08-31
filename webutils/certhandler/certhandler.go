package certhandler

import (
	"archive/zip"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/babbage88/go-infra/utils/env_helper"
)

var certPrefix string = "-----BEGIN CERTIFICATE-----"

var certSuffix string = "-----END CERTIFICATE-----"

var keyPrefix string = "-----BEGIN PRIVATE KEY-----"

var keySuffix string = "-----END PRIVATE KEY-----"

var certbotConfigDir string = ".certbot/config"

//authFile := env_helper.NewDotEnvSource(env_helper.WithVarName("JWT_KEY")).GetEnvVarValue()

type CertDnsRenewReq struct {
	AuthFile   string `json:"authFile"`
	DomainName string `json:"domainName"`
	Provider   string `json:"provider"`
	Email      string `json:"email"`
	ZipFiles   bool   `json:"zipFiles"`
}

type CertificateData struct {
	DomainName      string `json:"domainName"`
	CertPEM         string `json:"cert_pem"`
	ChainPEM        string `json:"chain_pem"`
	Fullchain       string `json:"fullchain_pem"`
	FullchainAndKey string `json:"fullchain_and_key"`
	PrivKey         string `json:"priv_key"`
	ZipDir          string
}

type Renewal interface {
	Renew() CertificateData
}

// CertRenewReq is the interface with the GetDomainName method.
type CertRenewReq interface {
	GetDomainName() string
}

func (c CertDnsRenewReq) GetDomainName() (string, error) {
	domain := strings.TrimPrefix(c.DomainName, "*.")
	if domain == "" {
		return "", errors.New("domain name cannot be empty")
	}
	return domain, nil
}

func (c CertDnsRenewReq) Renew(envars *env_helper.EnvVars) (CertificateData, error) {
	domname, err := c.GetDomainName()
	savedir := fmt.Sprint(domname, "/")

	authFile := envars.GetVarMapValue("CF_INI")
	fmt.Printf("authfil: %s", authFile)
	fmt.Printf("cert savedir value: %s", savedir)

	cmd := exec.Command("certbot",
		"certonly",
		"--dns-cloudflare",
		"--dns-cloudflare-credentials", authFile,
		"--dns-cloudflare-propagation-seconds", "60",
		"--email", c.Email,
		"--agree-tos",
		"--no-eff-email",
		"-d", c.DomainName,
		"--config-dir", ".certbot/config",
		"--logs-dir", ".certbot/logs",
		"--work-dir", ".certbot/work",
		"--cert-path", savedir, "--chain-path", savedir, "--fullchain-path", savedir, "--key-path", savedir,
	)

	var cert_info CertificateData

	slog.Info("Starting command to renew certificate", slog.String("Domain", c.DomainName), slog.String("DNS Provider", c.Provider))
	err = cmd.Run()

	if err != nil {
		slog.Error("Error executing renewal command.")
		return cert_info, err
	}
	live_dir := fmt.Sprint(certbotConfigDir, "/live/", savedir)

	cert_byte, _ := os.ReadFile(fmt.Sprint(live_dir, "cert.pem"))
	cert_str := string(cert_byte)

	chain_byte, _ := os.ReadFile(fmt.Sprint(live_dir, "chain.pem"))
	chain_str := string(chain_byte)

	fullchain_byte, _ := os.ReadFile(fmt.Sprint(live_dir, "fullchain.pem"))
	fullchain_str := string(fullchain_byte)

	privkey_byte, _ := os.ReadFile(fmt.Sprint(live_dir, "privkey.pem"))
	privkey_str := string(privkey_byte)

	fullchain_and_key := fmt.Sprint(fullchain_str, privkey_str)

	cert_info.CertPEM = cert_str
	cert_info.ChainPEM = chain_str
	cert_info.Fullchain = fullchain_str
	cert_info.PrivKey = privkey_str
	cert_info.DomainName = c.DomainName
	cert_info.FullchainAndKey = fullchain_and_key

	if c.ZipFiles {
		cert_info.ZipDir = fmt.Sprint(live_dir, "certs.zip")
		err := createZipFile(live_dir, cert_info)
		if err != nil {
			return cert_info, err
		}
	}

	return cert_info, err
}

// ReadAndTrimFile reads the content of a file, removes the PEM delimiters, and returns the trimmed content.
func ReadAndTrimFile(filename string, beginMarker string, endMarker string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	slog.Info("Parsing file conents ", slog.String("Filename", filename))
	contentStr := string(content)
	contentStr = strings.ReplaceAll(contentStr, beginMarker, "")
	contentStr = strings.ReplaceAll(contentStr, endMarker, "")
	slog.Info("Paring finished", slog.String("Content", contentStr))

	return strings.TrimSpace(contentStr), nil
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
		{"fullchainandkey.pem", certInfo.FullchainAndKey},
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
