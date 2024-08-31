package certhandler

import (
	"bytes"
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
}

type CertificateData struct {
	DomainName string `json:"domainName"`
	CertPEM    string `json:"cert_pem"`
	ChainPEM   string `json:"chain_pem"`
	Fullchain  string `json:"fullchain_pem"`
	PrivKey    string `json:"priv_key"`
	CmdOutput  string `json:"cmdoutput"`
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
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	slog.Info("Starting command to renew certificate", slog.String("Domain", c.DomainName), slog.String("DNS Provider", c.Provider))
	err = cmd.Run()

	cmdstring := fmt.Sprint("out:", outb.String(), "err:", errb.String())

	if err != nil {
		slog.Error("Error executing renewal command.")
		return cert_info, err
	}
	live_dir := fmt.Sprint(certbotConfigDir, "/live/", savedir)
	slog.Info("live_dir", slog.String("val", live_dir))

	cert_str, _ := ReadAndTrimFile(fmt.Sprint(live_dir, "cert.pem"), certPrefix, certSuffix)

	chain_str, _ := ReadAndTrimFile(fmt.Sprint(live_dir, "chain.pem"), certPrefix, certSuffix)

	fullchain_str, _ := ReadAndTrimFile(fmt.Sprint(live_dir, "fullchain.pem"), certPrefix, certSuffix)

	privkey_str, _ := ReadAndTrimFile(fmt.Sprint(live_dir, "privkey.pem"), keyPrefix, keySuffix)
	slog.Info("cert", slog.String("cert", cert_str))

	cert_info.CertPEM = cert_str
	cert_info.ChainPEM = chain_str
	cert_info.Fullchain = fullchain_str
	cert_info.PrivKey = privkey_str
	cert_info.DomainName = c.DomainName
	cert_info.CmdOutput = cmdstring

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
