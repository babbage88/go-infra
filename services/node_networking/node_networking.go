package node_networking

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/babbage88/go-infra/services/host_servers"
	"github.com/google/uuid"
	ping "github.com/prometheus-community/pro-bing"
)

// NetworkPingerImpl implements the NetworkPinger interface
type NetworkPingerImpl struct {
	hostServerProvider host_servers.HostServerProvider
}

// NewNetworkPinger creates a new NetworkPinger instance
func NewNetworkPinger(hostServerProvider host_servers.HostServerProvider) NetworkPinger {
	return &NetworkPingerImpl{
		hostServerProvider: hostServerProvider,
	}
}

// Ping pings an arbitrary IP address or hostname
func (n *NetworkPingerImpl) Ping(target string) PingResult {
	start := time.Now()

	pinger, err := ping.NewPinger(target)
	if err != nil {
		return PingResult{
			TargetHostName: target,
			Success:        false,
			Error:          fmt.Errorf("failed to create pinger: %w", err),
			Latency:        time.Since(start),
		}
	}

	// Set ping options
	pinger.Count = 1
	pinger.Timeout = 5 * time.Second
	pinger.SetPrivileged(false) // Use unprivileged ping

	// Run the ping
	err = pinger.Run()
	if err != nil {
		return PingResult{
			TargetHostName: target,
			Success:        false,
			Error:          fmt.Errorf("ping failed: %w", err),
			Latency:        time.Since(start),
		}
	}

	stats := pinger.Statistics()
	success := stats.PacketsRecv > 0

	return PingResult{
		TargetHostName: target,
		Success:        success,
		Error:          nil,
		Latency:        stats.AvgRtt,
	}
}

// PingHostServerNode pings a managed HostServer by its ID
func (n *NetworkPingerImpl) PingHostServerNode(hostServerNodeID uuid.UUID) PingResult {
	// Get the host server information
	hostServer, err := n.hostServerProvider.GetHostServer(context.Background(), hostServerNodeID)
	if err != nil {
		return PingResult{
			TargetHostId:   hostServerNodeID,
			TargetHostName: "",
			Success:        false,
			Error:          fmt.Errorf("failed to get host server: %w", err),
			Latency:        0,
		}
	}

	// Use the hostname for ping (fallback to IP if hostname is empty)
	target := hostServer.Hostname
	if target == "" {
		target = hostServer.IPAddress.String()
	}

	// Perform the ping
	result := n.Ping(target)
	result.TargetHostId = hostServerNodeID
	return result
}

// ProbeTCPPortByHostId probes a TCP port on a managed HostServer by its ID
func (n *NetworkPingerImpl) ProbeTCPPortByHostId(targetHostId uuid.UUID, port uint16) NetworkProbeResult {
	// Get the host server information
	hostServer, err := n.hostServerProvider.GetHostServer(context.Background(), targetHostId)
	if err != nil {
		return NetworkProbeResult{
			TargetHostId:   targetHostId,
			TargetHostName: "",
			TargetPort:     port,
			Success:        false,
			Error:          fmt.Errorf("failed to get host server: %w", err),
			Latency:        0,
		}
	}

	// Use the hostname for probe (fallback to IP if hostname is empty)
	target := hostServer.Hostname
	if target == "" {
		target = hostServer.IPAddress.String()
	}

	// Perform the TCP probe
	result := n.probeTCPPort(target, port)
	result.TargetHostId = targetHostId
	return result
}

// ProbeUDPPortByHostId probes a UDP port on a managed HostServer by its ID
func (n *NetworkPingerImpl) ProbeUDPPortByHostId(targetHostId uuid.UUID, port uint16) NetworkProbeResult {
	// Get the host server information
	hostServer, err := n.hostServerProvider.GetHostServer(context.Background(), targetHostId)
	if err != nil {
		return NetworkProbeResult{
			TargetHostId:   targetHostId,
			TargetHostName: "",
			TargetPort:     port,
			Success:        false,
			Error:          fmt.Errorf("failed to get host server: %w", err),
			Latency:        0,
		}
	}

	// Use the hostname for probe (fallback to IP if hostname is empty)
	target := hostServer.Hostname
	if target == "" {
		target = hostServer.IPAddress.String()
	}

	// Perform the UDP probe
	result := n.probeUDPPort(target, port)
	result.TargetHostId = targetHostId
	return result
}

// ProbeTCPPortByHostName probes a TCP port on a host by hostname
func (n *NetworkPingerImpl) ProbeTCPPortByHostName(targetHostName string, port uint16) NetworkProbeResult {
	result := n.probeTCPPort(targetHostName, port)
	result.TargetHostName = targetHostName
	return result
}

// ProbeUDPPortByHostName probes a UDP port on a host by hostname
func (n *NetworkPingerImpl) ProbeUDPPortByHostName(targetHostName string, port uint16) NetworkProbeResult {
	result := n.probeUDPPort(targetHostName, port)
	result.TargetHostName = targetHostName
	return result
}

// probeTCPPort performs a TCP port probe
func (n *NetworkPingerImpl) probeTCPPort(target string, port uint16) NetworkProbeResult {
	start := time.Now()

	address := fmt.Sprintf("%s:%d", target, port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return NetworkProbeResult{
			TargetHostName: target,
			TargetPort:     port,
			Success:        false,
			Error:          fmt.Errorf("TCP connection failed: %w", err),
			Latency:        time.Since(start),
		}
	}
	defer conn.Close()

	return NetworkProbeResult{
		TargetHostName: target,
		TargetPort:     port,
		Success:        true,
		Error:          nil,
		Latency:        time.Since(start),
	}
}

// probeUDPPort performs a UDP port probe
func (n *NetworkPingerImpl) probeUDPPort(target string, port uint16) NetworkProbeResult {
	start := time.Now()

	address := fmt.Sprintf("%s:%d", target, port)
	conn, err := net.DialTimeout("udp", address, 5*time.Second)
	if err != nil {
		return NetworkProbeResult{
			TargetHostName: target,
			TargetPort:     port,
			Success:        false,
			Error:          fmt.Errorf("UDP connection failed: %w", err),
			Latency:        time.Since(start),
		}
	}
	defer conn.Close()

	// For UDP, we can't reliably determine if the port is open
	// since UDP is connectionless. We'll consider it successful if we can establish a "connection"
	return NetworkProbeResult{
		TargetHostName: target,
		TargetPort:     port,
		Success:        true,
		Error:          nil,
		Latency:        time.Since(start),
	}
}
