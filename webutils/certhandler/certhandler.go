package certhandler

import (
	"bufio"
	"fmt"
	"log/slog"
	"os/exec"
)

type CertDnsRenewReq struct {
	AuthFile   string `json:"authFile"`
	DomainName string `json:"domainName"`
	Provider   string `json:"provider"`
}

type Renewal interface {
	Renew() []string
}

func (c CertDnsRenewReq) Renew() []string {
	cmd := exec.Command("certbot", "certonly", "--dns-cloudflare", "--dns-cloudflare-credentials", c.AuthFile, "--dns-cloudflare-propagation-seconds", "60", "-d", c.DomainName)

	slog.Info("Starting command to renew certificate", slog.String("Domain", c.DomainName), slog.String("DNS Provider", c.Provider))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		slog.Error("Error configuring StdoutPipe", slog.String("Error", err.Error()))
	}

	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		slog.Error("Error executing command", slog.String("Error", err.Error()))
	}

	// read command's stdout line by line
	in := bufio.NewScanner(stdout)
	var cmdoutput []string
	for in.Scan() {
		fmt.Printf(in.Text())
		cmdoutput = append(cmdoutput, in.Text())
	}

	return cmdoutput
}
