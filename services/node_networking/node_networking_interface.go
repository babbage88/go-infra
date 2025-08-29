package node_networking

import (
	"time"

	"github.com/google/uuid"
)

type NetworkProbeResult struct {
	TargetHostId   uuid.UUID
	TargetHostName string
	TargetPort     uint16
	Success        bool
	Error          error
	Latency        time.Duration
}

type PingResult struct {
	TargetHostId   uuid.UUID     `json:"targetHostId"`
	TargetHostName string        `json:"targetHostName"`
	IpAddrString   string        `json:"ipAddr,omitempty"`
	Success        bool          `json:"success"`
	Error          error         `json:"error"`
	AverageLatency time.Duration `json:"averageLatency"`
	PacketsSent    int           `json:"PacketsSent"`
	PacketsRecv    int           `json:"PacketsRecv"`
}

type NetworkPinger interface {
	Ping(remoteHost string) PingResult
	ArpPing(remoteHost string) PingResult
	PingHostServerNode(hostServerNodeID uuid.UUID) PingResult
	ProbeTCPPortByHostId(targetHostId uuid.UUID, port uint16) NetworkProbeResult
	ProbeUDPPortByHostId(targetHostId uuid.UUID, port uint16) NetworkProbeResult
	ProbeTCPPortByHostName(targetHostName string, port uint16) NetworkProbeResult
	ProbeUDPPortByHostName(targetHostName string, port uint16) NetworkProbeResult
}
