package proxmox

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	coredeploy "github.com/babbage88/infra-core/deployment"
)

const apiNodesPath = "/api2/json/nodes"

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	auth       authConfig
}

type authConfig struct {
	TokenID  string
	Secret   string
	Username string
	Password string
	UseToken bool
}

type apiError struct {
	Status int
	Body   string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("proxmox api error: status=%d body=%s", e.Status, e.Body)
}

func newClient(auth coredeploy.ProxmoxAuthOptions) (*Client, error) {
	hostURL := strings.TrimSpace(auth.HostURL)
	if hostURL == "" {
		return nil, fmt.Errorf("auth.host_url is required")
	}
	parsedURL, err := url.Parse(strings.TrimRight(hostURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("invalid auth.host_url: %w", err)
	}

	useToken := true
	if auth.UseToken != nil {
		useToken = *auth.UseToken
	}
	skipTLS := true
	if auth.SkipTLS != nil {
		skipTLS = *auth.SkipTLS
	}

	cfg := authConfig{UseToken: useToken}
	if useToken {
		cfg.TokenID = strings.TrimSpace(auth.APITokenID)
		cfg.Secret = strings.TrimSpace(auth.APISecret)
		if strings.TrimSpace(auth.APIToken) != "" {
			tokenID, secret, err := parseAPIToken(auth.APIToken)
			if err != nil {
				return nil, err
			}
			cfg.TokenID = tokenID
			cfg.Secret = secret
		}
		if cfg.TokenID == "" {
			return nil, fmt.Errorf("auth.api_token_id is required")
		}
		if cfg.Secret == "" {
			return nil, fmt.Errorf("auth.api_secret is required")
		}
	} else {
		cfg.Username = strings.TrimSpace(auth.Username)
		cfg.Password = auth.Password
		if cfg.Username == "" {
			return nil, fmt.Errorf("auth.username is required")
		}
		if strings.TrimSpace(cfg.Password) == "" {
			return nil, fmt.Errorf("auth.password is required")
		}
	}

	return &Client{
		baseURL: parsedURL,
		auth:    cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLS},
			},
		},
	}, nil
}

func (c *Client) do(ctx context.Context, method string, path string, body url.Values, out interface{}) error {
	fullURL := *c.baseURL
	fullURL.Path = strings.TrimRight(c.baseURL.Path, "/") + path

	var reader io.Reader
	if body != nil {
		reader = bytes.NewBufferString(body.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, fullURL.String(), reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if c.auth.UseToken {
		req.Header.Set("Authorization", "PVEAPIToken="+c.auth.TokenID+"="+c.auth.Secret)
	} else {
		return fmt.Errorf("password authentication is not supported yet")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read proxmox response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return &apiError{Status: resp.StatusCode, Body: string(bodyBytes)}
	}
	if out == nil || len(bodyBytes) == 0 {
		return nil
	}

	var wrapper struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &wrapper); err != nil {
		return fmt.Errorf("decode proxmox response: %w", err)
	}
	if len(wrapper.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(wrapper.Data, out); err != nil {
		return fmt.Errorf("decode proxmox data: %w", err)
	}
	return nil
}

func parseAPIToken(apiToken string) (string, string, error) {
	tokenID, secret, ok := strings.Cut(strings.TrimSpace(apiToken), "=")
	if !ok || tokenID == "" || secret == "" {
		return "", "", fmt.Errorf("invalid Proxmox API token format; expected USER@REALM!TOKENID=SECRET")
	}
	return tokenID, secret, nil
}
