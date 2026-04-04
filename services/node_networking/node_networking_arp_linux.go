//go:build linux

package node_networking

import (
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"time"
)

func (n *GoInfraRESTApiPinger) ArpPing(remoteHost string) PingResult {
	start := time.Now()
	ip := net.ParseIP(remoteHost)
	if ip == nil {
		return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("invalid IP"), AverageLatency: time.Since(start)}
	}

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

	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ARP)))
	if err != nil {
		return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("raw socket: %w", err), AverageLatency: time.Since(start)}
	}
	defer syscall.Close(fd)

	sll := syscall.SockaddrLinklayer{
		Protocol: htons(syscall.ETH_P_ARP),
		Ifindex:  iface.Index,
	}
	if err := syscall.Bind(fd, &sll); err != nil {
		return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("bind: %w", err), AverageLatency: time.Since(start)}
	}

	arpReq := &arpPacket{
		HType:  1,
		PType:  0x0800,
		HLen:   6,
		PLen:   4,
		OpCode: 1,
		SrcMAC: iface.HardwareAddr,
		SrcIP:  srcIP,
		DstMAC: net.HardwareAddr{0, 0, 0, 0, 0, 0},
		DstIP:  ip.To4(),
	}
	eth := &ethernetFrame{
		DstMAC:  net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		SrcMAC:  iface.HardwareAddr,
		EthType: 0x0806,
		Payload: arpReq.Marshal(),
	}
	packet := eth.Marshal()

	if err := syscall.Sendto(fd, packet, 0, &sll); err != nil {
		return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("send: %w", err), AverageLatency: time.Since(start)}
	}

	buf := make([]byte, 1500)
	syscall.SetsockoptTimeval(fd, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &syscall.Timeval{Sec: 2})
	_, from, err := syscall.Recvfrom(fd, buf, 0)
	if err != nil {
		return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("no reply: %w", err), AverageLatency: time.Since(start)}
	}

	_ = from
	if binary.BigEndian.Uint16(buf[12:14]) == 0x0806 {
		op := binary.BigEndian.Uint16(buf[20:22])
		if op == 2 {
			return PingResult{TargetHostName: remoteHost, Success: true, Error: nil, AverageLatency: time.Since(start)}
		}
	}

	return PingResult{TargetHostName: remoteHost, Success: false, Error: fmt.Errorf("no ARP reply received"), AverageLatency: time.Since(start)}
}
