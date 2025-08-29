package node_networking

import (
	"github.com/google/uuid"
)

// Ping Request/Response structs

// swagger:parameters pingHost
type PingRequestWrapper struct {
	// in:body
	Body PingRequest `json:"body"`
}

// swagger:model PingRequest
type PingRequest struct {
	// Target hostname or IP address to ping
	// required: true
	Target string `json:"target"`
}

// swagger:response PingResponse
type PingResponseWrapper struct {
	// in:body
	Body PingResponse `json:"body"`
}

// swagger:model PingResponse
type PingResponse struct {
	// ID of the target host server (if applicable)
	// required: false
	TargetHostId *uuid.UUID `json:"targetHostId,omitempty"`

	// Name of the target host
	// required: true
	TargetHostName string `json:"targetHostName"`

	// resolved ip address from hostname
	IpAddrString string `json:"ipAddr,omitempty"`

	// Whether the ping was successful
	// required: true
	Success bool `json:"success"`

	// Average Latency of the ping operation
	// required: true
	Latency string `json:"latency"`

	// Error message if the operation failed
	// required: false
	Error string `json:"error,omitempty"`

	// Total number of packets sent
	// required: false
	PacketsSent int `json:"PacketsSent,omitempty"`

	// Total number of recieved packets
	// required: false
	PacketsRecv int `json:"PacketsRecv,omitempty"`
}

// Ping Host Server Request/Response structs

// swagger:parameters pingHostServer
type PingHostServerRequestWrapper struct {
	// in:body
	Body PingHostServerRequest `json:"body"`
}

// swagger:model PingHostServerRequest
type PingHostServerRequest struct {
	// ID of the host server to ping
	// required: true
	HostServerID uuid.UUID `json:"hostServerId"`
}

// Network Probe Request/Response structs

// swagger:parameters probeTCPByHostname
type ProbeTCPByHostnameRequestWrapper struct {
	// in:body
	Body ProbeByHostnameRequest `json:"body"`
}

// swagger:parameters probeUDPByHostname
type ProbeUDPByHostnameRequestWrapper struct {
	// in:body
	Body ProbeByHostnameRequest `json:"body"`
}

// swagger:model ProbeByHostnameRequest
type ProbeByHostnameRequest struct {
	// Target hostname to probe
	// required: true
	TargetHostName string `json:"targetHostName"`

	// Port number to probe
	// required: true
	Port uint16 `json:"port"`
}

// swagger:parameters probeTCPByHostId
type ProbeTCPByHostIdRequestWrapper struct {
	// in:body
	Body ProbeByHostIdRequest `json:"body"`
}

// swagger:parameters probeUDPByHostId
type ProbeUDPByHostIdRequestWrapper struct {
	// in:body
	Body ProbeByHostIdRequest `json:"body"`
}

// swagger:model ProbeByHostIdRequest
type ProbeByHostIdRequest struct {
	// ID of the target host server
	// required: true
	TargetHostId uuid.UUID `json:"targetHostId"`

	// Port number to probe
	// required: true
	Port uint16 `json:"port"`
}

// swagger:response NetworkProbeResponse
type NetworkProbeResponseWrapper struct {
	// in:body
	Body NetworkProbeResponse `json:"body"`
}

// swagger:model NetworkProbeResponse
type NetworkProbeResponse struct {
	// ID of the target host server (if applicable)
	// required: false
	TargetHostId *uuid.UUID `json:"targetHostId,omitempty"`

	// Name of the target host
	// required: true
	TargetHostName string `json:"targetHostName"`

	// Port number that was probed
	// required: true
	TargetPort uint16 `json:"targetPort"`

	// Whether the probe was successful
	// required: true
	Success bool `json:"success"`

	// Latency of the probe operation
	// required: true
	Latency string `json:"latency"`

	// Error message if the operation failed
	// required: false
	Error string `json:"error,omitempty"`
}

// GET method request/response structs

// swagger:parameters pingHostGet
type PingHostGetRequestWrapper struct {
	// Target hostname or IP address to ping
	// in: path
	// required: true
	Target string `json:"target"`
}

// swagger:parameters probeTCPGet
type ProbeTCPGetRequestWrapper struct {
	// Target hostname to probe
	// in: path
	// required: true
	Target string `json:"target"`

	// Port number to probe
	// in: path
	// required: true
	Port string `json:"port"`
}

// swagger:parameters probeUDPGet
type ProbeUDPGetRequestWrapper struct {
	// Target hostname to probe
	// in: path
	// required: true
	Target string `json:"target"`

	// Port number to probe
	// in: path
	// required: true
	Port string `json:"port"`
}
