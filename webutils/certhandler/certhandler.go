package certhandler

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
)

type CertDnsRenewReq struct {
	AuthFile   string `json:"authFile"`
	DomainName string `json:"domainName"`
	Provider   string `json:"provider"`
	Email      string `json:"email"`
}

type Renewal interface {
	Renew() []string
}

func (c CertDnsRenewReq) Renew() []string {
	cmd := exec.Command("certbot",
		"certonly",
		"--dns-cloudflare",
		"--dns-cloudflare-credentials", c.AuthFile,
		"--dns-cloudflare-propagation-seconds", "60",
		"--email", c.Email,
		"--agree-tos",
		"--no-eff-email",
		"-d", c.DomainName,
		"--config-dir", ".certbot/config",
		"--logs-dir", ".certbot/logs",
		"--work-dir", ".certbot/work",
	)

	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	debugg := cmd.String()

	slog.Debug("Command being ran:", slog.String("Command", debugg))
	slog.Info("Starting command to renew certificate", slog.String("Domain", c.DomainName), slog.String("DNS Provider", c.Provider))
	// start the command after having set up the pipe
	if err := cmd.Run(); err != nil {
		slog.Error("Error executing command", slog.String("Error", err.Error()))
	}

	var cmdoutput []string
	cmdoutput = append(cmdoutput, outb.String(), errb.String())
	fmt.Println("out:", outb.String(), "err:", errb.String())

	return cmdoutput
}
