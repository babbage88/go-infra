package proxmox

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	coredeploy "github.com/babbage88/infra-core/deployment"
	coreproxmox "github.com/babbage88/infra-core/proxmox"
	"github.com/gorilla/websocket"
)

type vmConsoleSessionResponse struct {
	Port   int    `json:"port"`
	Ticket string `json:"ticket"`
	User   string `json:"user,omitempty"`
}

func VMConsoleSessionHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := parseListRequest(w, r)
		if !ok {
			return
		}

		vmid, err := parseVMIDPath(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, node, client, err := resolveConsoleClient(r.Context(), service, req, vmid)
		if err != nil {
			slog.Error("failed to resolve proxmox VM console session access", slog.Int("vmid", vmid), slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		proxyDetails, err := client.CreateQemuVNCProxy(r.Context(), node, vmid)
		if err != nil {
			slog.Error("failed to create proxmox VM console session", slog.Int("vmid", vmid), slog.String("node", node), slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, vmConsoleSessionResponse{
			Port:   proxyDetails.Port,
			Ticket: proxyDetails.Ticket,
			User:   proxyDetails.User,
		})
	}
}

func VMConsoleWebSocketHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := parseListRequest(w, r)
		if !ok {
			return
		}

		vmid, err := parseVMIDPath(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		auth, node, client, err := resolveConsoleClient(r.Context(), service, req, vmid)
		if err != nil {
			slog.Error("failed to resolve proxmox VM console access", slog.Int("vmid", vmid), slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		proxyDetails, err := vmConsoleProxyFromRequest(r, client, node, vmid)
		if err != nil {
			slog.Error("failed to create proxmox VM vncproxy", slog.Int("vmid", vmid), slog.String("node", node), slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		if err := bridgeProxmoxConsoleWebSocket(w, r, client, auth, node, "qemu", vmid, proxyDetails); err != nil {
			slog.Error("failed to bridge proxmox VM console websocket", slog.Int("vmid", vmid), slog.String("node", node), slog.String("error", err.Error()))
		}
	}
}

func vmConsoleProxyFromRequest(r *http.Request, client *coreproxmox.Client, node string, vmid int) (*coreproxmox.ConsoleProxyResponse, error) {
	portValue := strings.TrimSpace(r.URL.Query().Get("port"))
	ticket := strings.TrimSpace(r.URL.Query().Get("ticket"))
	if portValue == "" || ticket == "" {
		return client.CreateQemuVNCProxy(r.Context(), node, vmid)
	}

	port, err := strconv.Atoi(portValue)
	if err != nil || port <= 0 {
		return nil, fmt.Errorf("invalid console port")
	}

	return &coreproxmox.ConsoleProxyResponse{
		Port:   port,
		Ticket: ticket,
	}, nil
}

func LXCConsoleWebSocketHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := parseListRequest(w, r)
		if !ok {
			return
		}

		vmid, err := parseVMIDPath(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		auth, node, client, err := resolveConsoleClient(r.Context(), service, req, vmid)
		if err != nil {
			slog.Error("failed to resolve proxmox LXC console access", slog.Int("vmid", vmid), slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		proxyDetails, err := client.CreateLXCTermProxy(r.Context(), node, vmid, buildLXCTermProxyReferer(auth.HostURL, node, vmid))
		if err != nil {
			slog.Error("failed to create proxmox LXC termproxy", slog.Int("vmid", vmid), slog.String("node", node), slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		if err := bridgeProxmoxConsoleWebSocket(w, r, client, auth, node, "lxc", vmid, proxyDetails); err != nil {
			slog.Error("failed to bridge proxmox LXC console websocket", slog.Int("vmid", vmid), slog.String("node", node), slog.String("error", err.Error()))
		}
	}
}

func buildLXCTermProxyReferer(hostURL string, node string, vmid int) string {
	base, err := url.Parse(strings.TrimSpace(hostURL))
	if err != nil {
		return ""
	}

	query := url.Values{}
	query.Set("console", "lxc")
	query.Set("xtermjs", "1")
	query.Set("vmid", strconv.Itoa(vmid))
	query.Set("node", node)
	query.Set("cmd", "")

	base.Path = "/"
	base.RawQuery = query.Encode()
	return base.String()
}

func resolveConsoleClient(
	ctx context.Context,
	service *Service,
	req coredeploy.ProxmoxVMListRequest,
	vmid int,
) (coredeploy.ProxmoxAuthOptions, string, *coreproxmox.Client, error) {
	auth, _, node, err := service.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return coredeploy.ProxmoxAuthOptions{}, "", nil, err
	}
	if strings.TrimSpace(node) == "" {
		return coredeploy.ProxmoxAuthOptions{}, "", nil, fmt.Errorf("node is required")
	}
	if vmid <= 0 {
		return coredeploy.ProxmoxAuthOptions{}, "", nil, fmt.Errorf("vmid must be greater than zero")
	}
	client, err := newCoreClient(auth)
	if err != nil {
		return coredeploy.ProxmoxAuthOptions{}, "", nil, err
	}
	return auth, node, client, nil
}

func bridgeProxmoxConsoleWebSocket(
	w http.ResponseWriter,
	r *http.Request,
	client *coreproxmox.Client,
	auth coredeploy.ProxmoxAuthOptions,
	node string,
	workloadType string,
	vmid int,
	proxyDetails *coreproxmox.ConsoleProxyResponse,
) error {
	if proxyDetails == nil {
		http.Error(w, "console proxy details missing", http.StatusInternalServerError)
		return fmt.Errorf("console proxy details missing")
	}

	headers, err := client.AuthHeaders(r.Context(), false)
	if err != nil {
		http.Error(w, "failed to authorize proxmox console", http.StatusBadGateway)
		return err
	}

	hostURL, err := url.Parse(strings.TrimSpace(auth.HostURL))
	if err != nil {
		http.Error(w, "invalid proxmox host url", http.StatusInternalServerError)
		return err
	}
	wsScheme := "wss"
	if hostURL.Scheme == "http" {
		wsScheme = "ws"
	}
	wsURL := url.URL{
		Scheme: wsScheme,
		Host:   hostURL.Host,
		Path: fmt.Sprintf(
			"/api2/json/nodes/%s/%s/%d/vncwebsocket",
			url.PathEscape(node),
			workloadType,
			vmid,
		),
	}
	query := wsURL.Query()
	query.Set("port", fmt.Sprintf("%d", proxyDetails.Port))
	query.Set("vncticket", proxyDetails.Ticket)
	wsURL.RawQuery = query.Encode()

	dialer := client.WebsocketDialer()
	dialer.Subprotocols = websocket.Subprotocols(r)
	if dialer.TLSClientConfig == nil {
		skipTLS := true
		if auth.SkipTLS != nil {
			skipTLS = *auth.SkipTLS
		}
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: skipTLS}
	}

	upstreamConn, _, err := dialer.DialContext(r.Context(), wsURL.String(), headers)
	if err != nil {
		http.Error(w, "failed to connect to proxmox console websocket", http.StatusBadGateway)
		return err
	}
	defer upstreamConn.Close()

	if workloadType == "lxc" {
		if err := sendTermProxyAuth(upstreamConn, proxyDetails); err != nil {
			http.Error(w, "failed to authenticate proxmox container console", http.StatusBadGateway)
			return err
		}
	}

	upgrader := websocket.Upgrader{
		CheckOrigin:       func(_ *http.Request) bool { return true },
		EnableCompression: true,
		Subprotocols:      websocket.Subprotocols(r),
	}
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer clientConn.Close()

	errCh := make(chan error, 2)
	go relayWebSocket(errCh, clientConn, upstreamConn)
	go relayWebSocket(errCh, upstreamConn, clientConn)

	err = <-errCh
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
		return nil
	}
	if err == nil || strings.Contains(strings.ToLower(err.Error()), "closed") {
		return nil
	}
	return err
}

func sendTermProxyAuth(conn *websocket.Conn, proxyDetails *coreproxmox.ConsoleProxyResponse) error {
	if conn == nil {
		return fmt.Errorf("upstream console websocket missing")
	}
	if proxyDetails == nil {
		return fmt.Errorf("console proxy details missing")
	}

	user := strings.TrimSpace(proxyDetails.User)
	ticket := strings.TrimSpace(proxyDetails.Ticket)
	if user == "" || ticket == "" {
		return fmt.Errorf("termproxy auth requires user and ticket")
	}

	authLine := fmt.Sprintf("%s:%s\n", user, ticket)
	return conn.WriteMessage(websocket.TextMessage, []byte(authLine))
}

func relayWebSocket(errCh chan<- error, dst *websocket.Conn, src *websocket.Conn) {
	for {
		messageType, payload, err := src.ReadMessage()
		if err != nil {
			errCh <- err
			return
		}
		if err := dst.WriteMessage(messageType, payload); err != nil {
			errCh <- err
			return
		}
	}
}
