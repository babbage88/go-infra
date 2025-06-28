package ssh_connections

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/babbage88/goph/v2"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

func VerifyHost(host string, remote net.Addr, key ssh.PublicKey) error {
	// Get known_hosts path from env or default
	knownHostsPath := os.Getenv("SSH_KNOWN_HOSTS_PATH")
	if knownHostsPath == "" {
		knownHostsPath = os.ExpandEnv("ws_known_hosts")
	}

	hostFound, err := goph.CheckKnownHost(host, remote, key, knownHostsPath)
	if hostFound {
		if err != nil {
			// Key mismatch!
			return err // FAIL if mismatch
		}
		// Key matches
		return nil
	}
	// Host not found, automatically add to known_hosts (no prompt)
	return goph.AddKnownHost(host, remote, key, knownHostsPath)
}

func initializeSshClient(host string, user string, port uint, privateKey string, sshPassphrase string, timeout time.Duration) (*goph.Client, error) {
	var auth goph.Auth
	var err error
	slog.Info("host", "host", host)
	slog.Info("user", "user", user)
	slog.Info("port", "port", port)

	auth, err = goph.RawKey(privateKey, sshPassphrase)

	if err != nil {
		slog.Error("Failed to initialize SSH client", "error", err)
		return nil, err
	}

	client, err := goph.NewConn(&goph.Config{
		User:     user,
		Addr:     host,
		Port:     port,
		Auth:     auth,
		Callback: VerifyHost,
	})
	if err != nil {
		slog.Error("Failed to initialize SSH client", "error", err)
		return nil, err
	}
	// Defer closing the network connection.
	return client, err
}

func newGophClient(hostInfo *HostServerInfo, sshKey *SSHKeyInfo, config *SSHConfig) (*goph.Client, error) {
	if sshKey.Passphrase != "" {
		return initializeSshClient(hostInfo.IPAddress, sshKey.Username, uint(hostInfo.Port), sshKey.PrivateKey, sshKey.Passphrase, config.SSHTimeout)
	} else {
		return initializeSshClient(hostInfo.IPAddress, sshKey.Username, uint(hostInfo.Port), sshKey.PrivateKey, "", config.SSHTimeout)
	}
}

func (s *SSHSession) Connect(hostInfo *HostServerInfo, sshKey *SSHKeyInfo, config *SSHConfig, columns, rows int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	gophClient, err := newGophClient(hostInfo, sshKey, config)
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	// Create SSH session
	session, err := gophClient.NewSession()
	if err != nil {
		gophClient.Close()
		return fmt.Errorf("failed to create SSH session: %w", err)
	}

	// Request PTY (pseudo-terminal)
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// Use provided column and row sizes, with defaults if not specified
	if columns <= 0 {
		columns = 80
	}
	if rows <= 0 {
		rows = 24
	}

	// Request PTY with specified size and add detailed error logging
	if err := session.RequestPty("xterm", rows, columns, modes); err != nil {
		slog.Error("Failed to request PTY", "error", err, "columns", columns, "rows", rows)
		session.Close()
		gophClient.Close()
		return fmt.Errorf("failed to request PTY: %w", err)
	}

	s.SSHClient = gophClient.Client
	s.SSHSession = session

	// Log connection
	s.logConnection("connect", nil)

	return nil
}

// Set WebSocket connection
func (s *SSHSession) SetWebSocket(ws *websocket.Conn) {
	s.mu.Lock()
	s.WebSocket = ws
	s.mu.Unlock()
}

// Start bidirectional data transfer
func (s *SSHSession) StartDataTransfer() {
	slog.Info("StartDataTransfer: starting")
	if s.SSHSession == nil || s.WebSocket == nil {
		slog.Error("StartDataTransfer: missing SSHSession or WebSocket")
		return
	}

	// Set up bidirectional data transfer
	stdin, _ := s.SSHSession.StdinPipe()
	stdout, _ := s.SSHSession.StdoutPipe()
	stderr, _ := s.SSHSession.StderrPipe()

	// Start SSH session (interactive shell)
	err := s.SSHSession.Shell()
	if err != nil && err != io.EOF {
		slog.Error("Failed to start SSH shell", "error", err)
		return
	}

	// Start input/output goroutines (do NOT close session in these)
	go s.handleWebSocketInput(stdin)
	go s.handleSSHOutput(stdout, "data")
	go s.handleSSHOutput(stderr, "error")

	// Wait for the shell process to exit
	err = s.SSHSession.Wait()
	slog.Info("Shell process exited", "error", err)

	// Now close everything
	s.Close()
}

// Handle WebSocket input
func (s *SSHSession) handleWebSocketInput(stdin io.WriteCloser) {
	for {
		_, message, err := s.WebSocket.ReadMessage()
		if err != nil && err != io.EOF {
			slog.Error("handleWebSocketInput: WebSocket closed or error", "error", err.Error())
			return
		}

		s.updateActivity()

		// Parse message
		var wsMsg WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			slog.Error("Failed to parse WebSocket message", "error", err)
			continue
		}

		switch wsMsg.Type {
		case "input":
			if data, ok := wsMsg.Data.(string); ok {
				slog.Info("Writing to SSH stdin", "data", data)
				n, err := stdin.Write([]byte(data))
				if err != nil {
					if err == io.EOF {
						slog.Info("EOF on handleInput")
					}
					slog.Error("Failed to write to SSH stdin", "error", err)
				} else {
					slog.Info("Wrote bytes to SSH stdin", "count", n)
				}
			}
		case "resize":
			if resizeData, ok := wsMsg.Data.(map[string]interface{}); ok {
				cols := int(resizeData["cols"].(float64))
				rows := int(resizeData["rows"].(float64))
				s.SSHSession.WindowChange(rows, cols)
			}
		}
	}
}

// Handle SSH output
func (s *SSHSession) handleSSHOutput(reader io.Reader, msgType string) {
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			s.updateActivity()
			wsMsg := WebSocketMessage{
				Type: msgType,
				Data: string(buffer[:n]),
			}
			msgBytes, _ := json.Marshal(wsMsg)
			s.WebSocket.WriteMessage(websocket.TextMessage, msgBytes)
			slog.Info("SSH output", "type", msgType, "data", string(buffer[:n]))
			slog.Info("Sending to WebSocket", "data", string(buffer[:n]))
		}
		if err != nil {
			if err == io.EOF {
				slog.Info("EOF Reached")
				return
			}
			slog.Error("SSH read error", "type", msgType, "error", err)
			return
		}
	}
}

// Update last activity
func (s *SSHSession) updateActivity() {
	s.mu.Lock()
	s.LastActivity = time.Now()
	s.mu.Unlock()

	// Update database
	query := `UPDATE ssh_sessions SET last_activity = NOW() WHERE id = $1`
	s.dbtx.Exec(context.Background(), query, s.ID)
}

// Close session
func (s *SSHSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.SSHSession != nil {
		s.SSHSession.Close()
	}
	if s.SSHClient != nil {
		s.SSHClient.Close()
	}
	if s.WebSocket != nil {
		s.WebSocket.Close()
	}

	// Log disconnection
	s.logConnection("disconnect", nil)
}

// Log connection events
func (s *SSHSession) logConnection(action string, details map[string]interface{}) {
	detailsJSON, _ := json.Marshal(details)
	query := `
        INSERT INTO ssh_connection_logs (session_id, user_id, host_server_id, action, details)
        VALUES ($1, $2, $3, $4, $5)
    `
	s.dbtx.Exec(context.Background(), query, s.ID, s.UserID, s.HostServerID, action, detailsJSON)
}

// GetHostKeyCallback returns a host key callback for SSH connections
func GetHostKeyCallback(knownHostsPath string) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// Read known_hosts file
		file, err := os.Open(knownHostsPath)
		if err != nil {
			// If file doesn't exist, create it and accept the key
			return acceptAndSaveHostKey(knownHostsPath, hostname, key)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Parse known_hosts line: hostname,ip ssh-rsa key
			parts := strings.Fields(line)
			if len(parts) >= 3 && (parts[0] == hostname || parts[0] == remote.String()) {
				// Verify the key matches
				if strings.TrimSpace(parts[2]) == base64.StdEncoding.EncodeToString(key.Marshal()) {
					return nil
				}
			}
		}

		// Key not found, accept and save it
		return acceptAndSaveHostKey(knownHostsPath, hostname, key)
	}
}

func acceptAndSaveHostKey(knownHostsPath, hostname string, key ssh.PublicKey) error {
	// Create directory if it doesn't exist
	dir := strings.TrimSuffix(knownHostsPath, "/known_hosts")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// Append the new host key
	file, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open known_hosts: %w", err)
	}
	defer file.Close()

	keyLine := fmt.Sprintf("%s ssh-rsa %s\n", hostname, base64.StdEncoding.EncodeToString(key.Marshal()))
	if _, err := file.WriteString(keyLine); err != nil {
		return fmt.Errorf("failed to write to known_hosts: %w", err)
	}

	return nil
}
