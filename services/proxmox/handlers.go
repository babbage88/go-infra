package proxmox

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	coredeploy "github.com/babbage88/infra-core/deployment"
)

// swagger:route GET /api/v1/proxmox/vm Proxmox ListProxmoxVMs
// List Proxmox QEMU VMs.
// responses:
//
//	200: ProxmoxVMListResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func ListVMsHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := coredeploy.ProxmoxVMListRequest{}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		if node := r.URL.Query().Get("node"); node != "" {
			req.Node = node
		}
		if full := r.URL.Query().Get("full"); full != "" {
			value, err := strconv.ParseBool(full)
			if err != nil {
				http.Error(w, "full must be a boolean", http.StatusBadRequest)
				return
			}
			req.Full = &value
		}

		result, err := service.ListVMs(r.Context(), req)
		if err != nil {
			slog.Error("failed to list proxmox VMs", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route POST /api/v1/proxmox/vm/{vmid}/start Proxmox StartProxmoxVM
// Start a Proxmox QEMU VM.
// responses:
//
//	200: ProxmoxVMStartResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func StartVMHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vmid, err := strconv.Atoi(r.PathValue("vmid"))
		if err != nil || vmid <= 0 {
			http.Error(w, "vmid path parameter must be a positive integer", http.StatusBadRequest)
			return
		}

		req := coredeploy.ProxmoxVMStartRequest{VMID: vmid}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		req.VMID = vmid

		result, err := service.StartVM(r.Context(), req)
		if err != nil {
			slog.Error("failed to start proxmox VM", slog.Int("vmid", vmid), slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
