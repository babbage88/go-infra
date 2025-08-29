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
	TargetHostId   uuid.UUID
	TargetHostName string
	Success        bool
	Error          error
	Latency        time.Duration
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
