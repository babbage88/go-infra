package node_networking

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"time"

	"github.com/babbage88/go-infra/services/host_servers"
	"github.com/google/uuid"
	ping "github.com/prometheus-community/pro-bing"
)

// GoInfraRESTApiPinger implements the NetworkPinger interface
type GoInfraRESTApiPinger struct {
	hostServerProvider host_servers.HostServerProvider
}

// NewNetworkPinger creates a new NetworkPinger instance
func NewNetworkPinger(hostServerProvider host_servers.HostServerProvider) NetworkPinger {
	return &GoInfraRESTApiPinger{
		hostServerProvider: hostServerProvider,
	}
}

func (n *GoInfraRESTApiPinger) ArpPing(remoteHost string) PingResult {
	start := time.Now()
	ip := net.ParseIP(remoteHost)
	if ip == nil {
		return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("invalid IP"), AverageLatency: time.Since(start)}
	}

	// choose an interface (first non-loopback)
	ifaces, _ := net.Interfaces()
	var iface *net.Interface
	for _, i := range ifaces {
		if i.Flags&net.FlagUp != 0 && i.Flags&net.FlagLoopback == 0 && len(i.HardwareAddr) == 6 {
			iface = &i
			break
		}
	}
	if iface == nil {
		return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("no usable interface"), AverageLatency: time.Since(start)}
	}

	// get source IPv4 from iface
	addrs, _ := iface.Addrs()
	var srcIP net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			srcIP = ipnet.IP.To4()
			break
		}
	}
	if srcIP == nil {
		return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("no IPv4 on iface"), AverageLatency: time.Since(start)}
	}

	// open raw socket
	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ARP)))
	if err != nil {
		return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("raw socket: %w", err), AverageLatency: time.Since(start)}
	}
	defer syscall.Close(fd)

	// bind to interface
	sll := syscall.SockaddrLinklayer{
		Protocol: htons(syscall.ETH_P_ARP),
		Ifindex:  iface.Index,
	}
	if err := syscall.Bind(fd, &sll); err != nil {
		return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("bind: %w", err), AverageLatency: time.Since(start)}
	}

	// build ARP request
	arpReq := &arpPacket{
		HType:  1,      // Ethernet
		PType:  0x0800, // IPv4
		HLen:   6,
		PLen:   4,
		OpCode: 1, // request
		SrcMAC: iface.HardwareAddr,
		SrcIP:  srcIP,
		DstMAC: net.HardwareAddr{0, 0, 0, 0, 0, 0},
		DstIP:  ip.To4(),
	}
	eth := &ethernetFrame{
		DstMAC:  net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		SrcMAC:  iface.HardwareAddr,
		EthType: 0x0806, // ARP
		Payload: arpReq.Marshal(),
	}
	packet := eth.Marshal()

	// send
	if err := syscall.Sendto(fd, packet, 0, &sll); err != nil {
		return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("send: %w", err), AverageLatency: time.Since(start)}
	}

	// wait for reply
	buf := make([]byte, 1500)
	syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &syscall.Timeval{Sec: 2})
	_, from, err := syscall.Recvfrom(fd, buf, 0)
	if err != nil {
		return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("no reply: %w", err), AverageLatency: time.Since(start)}
	}

	// confirm it's ARP reply
	_ = from                                           // ignore linklayer details
	if binary.BigEndian.Uint16(buf[12:14]) == 0x0806 { // Ethertype = ARP
		op := binary.BigEndian.Uint16(buf[20:22])
		if op == 2 { // reply
			return PingResult{TargetHostName: remoteHost, Success: true, Error: nil, AverageLatency: time.Since(start)}
		}
	}

	return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("no ARP reply received"), AverageLatency: time.Since(start)}
}

// Ping pings an arbitrary IP address or hostname
func (n *GoInfraRESTApiPinger) Ping(remoteHost string) PingResult {
	start := time.Now()

	pinger, err := ping.NewPinger(remoteHost)
	if err != nil {
		return PingResult{
			TargetHostName: remoteHost,
			Success:        false,
			Error:          fmt.Errorf("failed to create pinger: %w", err),
			AverageLatency: time.Since(start),
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
			TargetHostName: remoteHost,
			IpAddrString:   "",
			Success:        false,
			Error:          fmt.Errorf("ping failed: %w", err),
			AverageLatency: time.Since(start),
		}
	}

	stats := pinger.Statistics()
	success := stats.PacketsRecv > 0

	return PingResult{
		TargetHostName: remoteHost,
		IpAddrString:   stats.IPAddr.String(),
		Success:        success,
		Error:          nil,
		AverageLatency: stats.AvgRtt,
		PacketsSent:    stats.PacketsSent,
		PacketsRecv:    stats.PacketsRecv,
	}
}

// PingHostServerNode pings a managed HostServer by its ID
func (n *GoInfraRESTApiPinger) PingHostServerNode(hostServerNodeID uuid.UUID) PingResult {
	// Get the host server information
	hostServer, err := n.hostServerProvider.GetHostServer(context.Background(), hostServerNodeID)
	if err != nil {
		return PingResult{
			TargetHostId:   hostServerNodeID,
			TargetHostName: "",
			Success:        false,
			Error:          fmt.Errorf("failed to get host server: %w", err),
			AverageLatency: 0,
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
func (n *GoInfraRESTApiPinger) ProbeTCPPortByHostId(targetHostId uuid.UUID, port uint16) NetworkProbeResult {
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
func (n *GoInfraRESTApiPinger) ProbeUDPPortByHostId(targetHostId uuid.UUID, port uint16) NetworkProbeResult {
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
func (n *GoInfraRESTApiPinger) ProbeTCPPortByHostName(targetHostName string, port uint16) NetworkProbeResult {
	result := n.probeTCPPort(targetHostName, port)
	result.TargetHostName = targetHostName
	return result
}

// ProbeUDPPortByHostName probes a UDP port on a host by hostname
func (n *GoInfraRESTApiPinger) ProbeUDPPortByHostName(targetHostName string, port uint16) NetworkProbeResult {
	result := n.probeUDPPort(targetHostName, port)
	result.TargetHostName = targetHostName
	return result
}

// probeTCPPort performs a TCP port probe
func (n *GoInfraRESTApiPinger) probeTCPPort(target string, port uint16) NetworkProbeResult {
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
func (n *GoInfraRESTApiPinger) probeUDPPort(target string, port uint16) NetworkProbeResult {
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
