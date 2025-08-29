package node_networking

import (
	"encoding/json"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/google/uuid"
)

// swagger:route POST /network/ping network-ping pingHost
// Ping an arbitrary hostname or IP address.
// responses:
//
//	200: PingResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func PingHandler(pinger NetworkPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Target == "" {
			http.Error(w, "Target hostname or IP is required", http.StatusBadRequest)
			return
		}

		// Perform the ping
		result := pinger.Ping(req.Target)

		// Prepare response
		resp := PingResponse{
			TargetHostName: result.TargetHostName,
			IpAddrString:   result.IpAddrString,
			Success:        result.Success,
			Latency:        result.AverageLatency.String(),
			PacketsSent:    result.PacketsSent,
			PacketsRecv:    result.PacketsRecv,
		}

		// Set TargetHostId if it's not zero
		if result.TargetHostId != uuid.Nil {
			resp.TargetHostId = &result.TargetHostId
		}

		if result.Error != nil {
			resp.Error = result.Error.Error()
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// swagger:route POST /network/ping-host-server network-ping pingHostServer
// Ping a managed host server by its ID.
// responses:
//
//	200: PingResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	404: description:Host server not found
//	500: description:Internal Server Error
func PingHostServerHandler(pinger NetworkPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req PingHostServerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.HostServerID == uuid.Nil {
			http.Error(w, "Host server ID is required", http.StatusBadRequest)
			return
		}

		// Perform the ping
		result := pinger.PingHostServerNode(req.HostServerID)

		// Prepare response
		resp := PingResponse{
			TargetHostId:   &req.HostServerID,
			TargetHostName: result.TargetHostName,
			IpAddrString:   result.IpAddrString,
			Success:        result.Success,
			Latency:        result.AverageLatency.String(),
			PacketsSent:    result.PacketsSent,
			PacketsRecv:    result.PacketsRecv,
		}

		// Set TargetHostId if it's not zero
		if result.TargetHostId != uuid.Nil {
			resp.TargetHostId = &result.TargetHostId
		}

		if result.Error != nil {
			resp.Error = result.Error.Error()
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// swagger:route POST /network/probe-tcp-hostname network-probe probeTCPByHostname
// Probe a TCP port on a host by hostname.
// responses:
//
//	200: NetworkProbeResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func ProbeTCPByHostnameHandler(pinger NetworkPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ProbeByHostnameRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.TargetHostName == "" {
			http.Error(w, "Target hostname is required", http.StatusBadRequest)
			return
		}
		if req.Port == 0 {
			http.Error(w, "Port is required", http.StatusBadRequest)
			return
		}

		// Perform the TCP probe
		result := pinger.ProbeTCPPortByHostName(req.TargetHostName, req.Port)

		// Prepare response
		resp := NetworkProbeResponse{
			TargetHostName: result.TargetHostName,
			TargetPort:     result.TargetPort,
			Success:        result.Success,
			Latency:        result.Latency.String(),
		}

		// Set TargetHostId if it's not zero
		if result.TargetHostId != uuid.Nil {
			resp.TargetHostId = &result.TargetHostId
		}

		if result.Error != nil {
			resp.Error = result.Error.Error()
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// swagger:route POST /network/probe-udp-hostname network-probe probeUDPByHostname
// Probe a UDP port on a host by hostname.
// responses:
//
//	200: NetworkProbeResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func ProbeUDPByHostnameHandler(pinger NetworkPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ProbeByHostnameRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.TargetHostName == "" {
			http.Error(w, "Target hostname is required", http.StatusBadRequest)
			return
		}
		if req.Port == 0 {
			http.Error(w, "Port is required", http.StatusBadRequest)
			return
		}

		// Perform the UDP probe
		result := pinger.ProbeUDPPortByHostName(req.TargetHostName, req.Port)

		// Prepare response
		resp := NetworkProbeResponse{
			TargetHostName: result.TargetHostName,
			TargetPort:     result.TargetPort,
			Success:        result.Success,
			Latency:        result.Latency.String(),
		}

		// Set TargetHostId if it's not zero
		if result.TargetHostId != uuid.Nil {
			resp.TargetHostId = &result.TargetHostId
		}

		if result.Error != nil {
			resp.Error = result.Error.Error()
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// swagger:route POST /network/probe-tcp-host-id network-probe probeTCPByHostId
// Probe a TCP port on a managed host server by its ID.
// responses:
//
//	200: NetworkProbeResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	404: description:Host server not found
//	500: description:Internal Server Error
func ProbeTCPByHostIdHandler(pinger NetworkPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ProbeByHostIdRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.TargetHostId == uuid.Nil {
			http.Error(w, "Target host ID is required", http.StatusBadRequest)
			return
		}
		if req.Port == 0 {
			http.Error(w, "Port is required", http.StatusBadRequest)
			return
		}

		// Perform the TCP probe
		result := pinger.ProbeTCPPortByHostId(req.TargetHostId, req.Port)

		// Prepare response
		resp := NetworkProbeResponse{
			TargetHostId:   &req.TargetHostId,
			TargetHostName: result.TargetHostName,
			TargetPort:     result.TargetPort,
			Success:        result.Success,
			Latency:        result.Latency.String(),
		}

		// Set TargetHostId if it's not zero
		if result.TargetHostId != uuid.Nil {
			resp.TargetHostId = &result.TargetHostId
		}

		if result.Error != nil {
			resp.Error = result.Error.Error()
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// swagger:route POST /network/probe-udp-host-id network-probe probeUDPByHostId
// Probe a UDP port on a managed host server by its ID.
// responses:
//
//	200: NetworkProbeResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	404: description:Host server not found
//	500: description:Internal Server Error
func ProbeUDPByHostIdHandler(pinger NetworkPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ProbeByHostIdRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.TargetHostId == uuid.Nil {
			http.Error(w, "Target host ID is required", http.StatusBadRequest)
			return
		}
		if req.Port == 0 {
			http.Error(w, "Port is required", http.StatusBadRequest)
			return
		}

		// Perform the UDP probe
		result := pinger.ProbeUDPPortByHostId(req.TargetHostId, req.Port)

		// Prepare response
		resp := NetworkProbeResponse{
			TargetHostId:   &req.TargetHostId,
			TargetHostName: result.TargetHostName,
			TargetPort:     result.TargetPort,
			Success:        result.Success,
			Latency:        result.Latency.String(),
		}

		// Set TargetHostId if it's not zero
		if result.TargetHostId != uuid.Nil {
			resp.TargetHostId = &result.TargetHostId
		}

		if result.Error != nil {
			resp.Error = result.Error.Error()
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// Convenience handler for ping operations with query parameters
// swagger:route GET /network/ping/{target} network-ping pingHostGet
// Ping an arbitrary hostname or IP address using GET method.
// responses:
//
//	200: PingResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func PingGetHandler(pinger NetworkPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get target from URL path
		target := r.PathValue("target")
		if target == "" {
			http.Error(w, "Target hostname or IP is required", http.StatusBadRequest)
			return
		}

		// Perform the ping
		result := pinger.Ping(target)

		// Prepare response
		resp := PingResponse{
			TargetHostName: result.TargetHostName,
			IpAddrString:   result.IpAddrString,
			Success:        result.Success,
			Latency:        result.AverageLatency.String(),
			PacketsSent:    result.PacketsSent,
			PacketsRecv:    result.PacketsRecv,
		}

		// Set TargetHostId if it's not zero
		if result.TargetHostId != uuid.Nil {
			resp.TargetHostId = &result.TargetHostId
		}

		if result.Error != nil {
			resp.Error = result.Error.Error()
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// Convenience handler for TCP probe operations with query parameters
// swagger:route GET /network/probe-tcp/{target}/{port} network-probe probeTCPGet
// Probe a TCP port on a host using GET method.
// responses:
//
//	200: NetworkProbeResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func ProbeTCPGetHandler(pinger NetworkPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get target and port from URL path
		target := r.PathValue("target")
		portStr := r.PathValue("port")

		if target == "" {
			http.Error(w, "Target hostname is required", http.StatusBadRequest)
			return
		}
		if portStr == "" {
			http.Error(w, "Port is required", http.StatusBadRequest)
			return
		}

		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			http.Error(w, "Invalid port number", http.StatusBadRequest)
			return
		}

		// Perform the TCP probe
		result := pinger.ProbeTCPPortByHostName(target, uint16(port))

		// Prepare response
		resp := NetworkProbeResponse{
			TargetHostName: result.TargetHostName,
			TargetPort:     result.TargetPort,
			Success:        result.Success,
			Latency:        result.Latency.String(),
		}

		// Set TargetHostId if it's not zero
		if result.TargetHostId != uuid.Nil {
			resp.TargetHostId = &result.TargetHostId
		}

		if result.Error != nil {
			resp.Error = result.Error.Error()
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// Convenience handler for UDP probe operations with query parameters
// swagger:route GET /network/probe-udp/{target}/{port} network-probe probeUDPGet
// Probe a UDP port on a host using GET method.
// responses:
//
//	200: NetworkProbeResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func ProbeUDPGetHandler(pinger NetworkPinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get target and port from URL path
		target := r.PathValue("target")
		portStr := r.PathValue("port")

		if target == "" {
			http.Error(w, "Target hostname is required", http.StatusBadRequest)
			return
		}
		if portStr == "" {
			http.Error(w, "Port is required", http.StatusBadRequest)
			return
		}

		port, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			http.Error(w, "Invalid port number", http.StatusBadRequest)
			return
		}

		// Perform the UDP probe
		result := pinger.ProbeUDPPortByHostName(target, uint16(port))

		// Prepare response
		resp := NetworkProbeResponse{
			TargetHostName: result.TargetHostName,
			TargetPort:     result.TargetPort,
			Success:        result.Success,
			Latency:        result.Latency.String(),
		}

		// Set TargetHostId if it's not zero
		if result.TargetHostId != uuid.Nil {
			resp.TargetHostId = &result.TargetHostId
		}

		if result.Error != nil {
			resp.Error = result.Error.Error()
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
