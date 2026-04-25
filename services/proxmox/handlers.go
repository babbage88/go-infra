package proxmox

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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
		req, ok := parseListRequest(w, r)
		if !ok {
			return
		}

		result, err := service.ListVMs(r.Context(), req)
		if err != nil {
			slog.Error("failed to list proxmox VMs", slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route GET /api/v1/proxmox/container Proxmox ListProxmoxContainers
// List Proxmox LXC containers.
// responses:
//
//	200: ProxmoxContainerListResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func ListContainersHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := parseListRequest(w, r)
		if !ok {
			return
		}

		result, err := service.ListContainers(r.Context(), req)
		if err != nil {
			slog.Error("failed to list proxmox containers", slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route GET /api/v1/proxmox/workload Proxmox ListProxmoxWorkloads
// List all Proxmox QEMU VMs and LXC containers for a node.
// responses:
//
//	200: ProxmoxWorkloadInventoryResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func ListWorkloadsHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, ok := parseListRequest(w, r)
		if !ok {
			return
		}

		result, err := service.ListWorkloads(r.Context(), req)
		if err != nil {
			slog.Error("failed to list proxmox workloads", slog.String("error", err.Error()))
			writeServiceError(w, err)
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
			writeServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route POST /api/v1/proxmox/lxc Proxmox CreateProxmoxLXC
// Create a Proxmox LXC container.
// responses:
//
//	200: ProxmoxLXCResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func CreateLXCHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := coredeploy.ProxmoxLXCRequest{}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}

		result, err := service.CreateLXC(r.Context(), req)
		if err != nil {
			slog.Error("failed to create proxmox LXC", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route POST /api/v1/proxmox/vm Proxmox CreateProxmoxVM
// Create a Proxmox VM from a template.
// responses:
//
//	200: ProxmoxVMCreateResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func CreateVMHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := coredeploy.ProxmoxVMCreateRequest{}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}

		result, err := service.CreateVM(r.Context(), req)
		if err != nil {
			slog.Error("failed to create proxmox VM", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route POST /api/v1/proxmox/vm/template Proxmox CreateProxmoxVMTemplate
// Create a Proxmox VM template from a cloud image.
// responses:
//
//	200: ProxmoxVMTemplateResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func CreateVMTemplateHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := coredeploy.ProxmoxVMTemplateRequest{}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}

		result, err := service.CreateVMTemplate(r.Context(), req)
		if err != nil {
			slog.Error("failed to create proxmox VM template", slog.String("error", err.Error()))
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

func writeServiceError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	if strings.Contains(err.Error(), "is required") || strings.Contains(err.Error(), "must be greater than zero") {
		status = http.StatusBadRequest
	}
	http.Error(w, err.Error(), status)
}

func parseListRequest(w http.ResponseWriter, r *http.Request) (coredeploy.ProxmoxVMListRequest, bool) {
	req := coredeploy.ProxmoxVMListRequest{}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return req, false
		}
	}
	if node := r.URL.Query().Get("node"); node != "" {
		req.Node = node
	}
	if full := r.URL.Query().Get("full"); full != "" {
		value, err := strconv.ParseBool(full)
		if err != nil {
			http.Error(w, "full must be a boolean", http.StatusBadRequest)
			return req, false
		}
		req.Full = &value
	}
	return req, true
}
