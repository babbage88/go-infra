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

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

func (s *SSHSession) Connect(hostInfo *HostServerInfo, sshKey *SSHKeyInfo, config *SSHConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse SSH private key
	signer, err := ssh.ParsePrivateKey([]byte(sshKey.PrivateKey))
	if err != nil {
		return fmt.Errorf("failed to parse SSH key: %w", err)
	}

	// SSH client configuration
	sshConfig := &ssh.ClientConfig{
		User: sshKey.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: GetHostKeyCallback(config.KnownHostsPath),
		Timeout:         config.SSHTimeout,
	}

	// Connect to SSH server
	port := hostInfo.Port
	if port == 0 {
		port = 22
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", hostInfo.IPAddress, port), sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	// Create SSH session
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to create SSH session: %w", err)
	}

	// Request PTY (pseudo-terminal)
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		session.Close()
		client.Close()
		return fmt.Errorf("failed to request PTY: %w", err)
	}

	s.SSHClient = client
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
	if s.SSHSession == nil || s.WebSocket == nil {
		return
	}

	// Set up bidirectional data transfer
	stdin, _ := s.SSHSession.StdinPipe()
	stdout, _ := s.SSHSession.StdoutPipe()
	stderr, _ := s.SSHSession.StderrPipe()

	// Start SSH session
	if err := s.SSHSession.Shell(); err != nil {
		slog.Error("Failed to start SSH shell", "error", err)
		return
	}

	// Handle WebSocket messages (input from browser)
	go s.handleWebSocketInput(stdin)

	// Handle SSH output (to browser)
	go s.handleSSHOutput(stdout, "data")

	// Handle SSH errors (to browser)
	go s.handleSSHOutput(stderr, "error")
}

// Handle WebSocket input
func (s *SSHSession) handleWebSocketInput(stdin io.WriteCloser) {
	defer s.Close()

	for {
		_, message, err := s.WebSocket.ReadMessage()
		if err != nil {
			slog.Error("WebSocket read error", "error", err)
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
				stdin.Write([]byte(data))
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
	defer s.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			slog.Error("SSH read error", "type", msgType, "error", err)
			return
		}
		if n > 0 {
			s.updateActivity()

			wsMsg := WebSocketMessage{
				Type: msgType,
				Data: string(buffer[:n]),
			}
			msgBytes, _ := json.Marshal(wsMsg)
			s.WebSocket.WriteMessage(websocket.TextMessage, msgBytes)
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
