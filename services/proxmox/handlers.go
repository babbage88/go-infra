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
	coreproxmox "github.com/babbage88/infra-core/proxmox"
	"github.com/google/uuid"
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

		body := ProxmoxVMStartRequest{VMID: vmid}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		req := coredeploy.ProxmoxVMStartRequest{
			HostServerID:    body.HostServerID,
			ProxmoxSecretID: body.ProxmoxSecretID,
			Node:            body.Node,
			VMID:            vmid,
		}

		result, err := service.StartVM(r.Context(), req)
		if err != nil {
			slog.Error("failed to start proxmox VM", slog.Int("vmid", vmid), slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route POST /api/v1/proxmox/vm/{vmid}/stop Proxmox StopProxmoxVM
// Stop a Proxmox QEMU VM.
// responses:
//
//	200: ProxmoxVMStartResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func StopVMHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vmid, err := strconv.Atoi(r.PathValue("vmid"))
		if err != nil || vmid <= 0 {
			http.Error(w, "vmid path parameter must be a positive integer", http.StatusBadRequest)
			return
		}

		body := ProxmoxVMStartRequest{VMID: vmid}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		req := coredeploy.ProxmoxVMStartRequest{
			HostServerID:    body.HostServerID,
			ProxmoxSecretID: body.ProxmoxSecretID,
			Node:            body.Node,
			VMID:            vmid,
		}

		result, err := service.StopVM(r.Context(), req)
		if err != nil {
			slog.Error("failed to stop proxmox VM", slog.Int("vmid", vmid), slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route POST /api/v1/proxmox/container/{vmid}/start Proxmox StartProxmoxContainer
// Start a Proxmox LXC container.
// responses:
//
//	200: ProxmoxVMStartResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func StartContainerHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vmid, err := strconv.Atoi(r.PathValue("vmid"))
		if err != nil || vmid <= 0 {
			http.Error(w, "vmid path parameter must be a positive integer", http.StatusBadRequest)
			return
		}

		body := ProxmoxVMStartRequest{VMID: vmid}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		req := coredeploy.ProxmoxVMStartRequest{
			HostServerID:    body.HostServerID,
			ProxmoxSecretID: body.ProxmoxSecretID,
			Node:            body.Node,
			VMID:            vmid,
		}

		result, err := service.StartContainer(r.Context(), req)
		if err != nil {
			slog.Error("failed to start proxmox container", slog.Int("vmid", vmid), slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route POST /api/v1/proxmox/container/{vmid}/stop Proxmox StopProxmoxContainer
// Stop a Proxmox LXC container.
// responses:
//
//	200: ProxmoxVMStartResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func StopContainerHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vmid, err := strconv.Atoi(r.PathValue("vmid"))
		if err != nil || vmid <= 0 {
			http.Error(w, "vmid path parameter must be a positive integer", http.StatusBadRequest)
			return
		}

		body := ProxmoxVMStartRequest{VMID: vmid}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		req := coredeploy.ProxmoxVMStartRequest{
			HostServerID:    body.HostServerID,
			ProxmoxSecretID: body.ProxmoxSecretID,
			Node:            body.Node,
			VMID:            vmid,
		}

		result, err := service.StopContainer(r.Context(), req)
		if err != nil {
			slog.Error("failed to stop proxmox container", slog.Int("vmid", vmid), slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route DELETE /api/v1/proxmox/vm/{vmid} Proxmox DeleteProxmoxVM
// Delete a Proxmox QEMU VM.
// responses:
//
//	200: ProxmoxVMStartResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func DeleteVMHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vmid, err := strconv.Atoi(r.PathValue("vmid"))
		if err != nil || vmid <= 0 {
			http.Error(w, "vmid path parameter must be a positive integer", http.StatusBadRequest)
			return
		}

		body := ProxmoxVMStartRequest{VMID: vmid}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		req := coredeploy.ProxmoxVMStartRequest{
			HostServerID:    body.HostServerID,
			ProxmoxSecretID: body.ProxmoxSecretID,
			Node:            body.Node,
			VMID:            vmid,
		}

		result, err := service.DeleteVM(r.Context(), req)
		if err != nil {
			slog.Error("failed to delete proxmox VM", slog.Int("vmid", vmid), slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route DELETE /api/v1/proxmox/container/{vmid} Proxmox DeleteProxmoxContainer
// Delete a Proxmox LXC container.
// responses:
//
//	200: ProxmoxVMStartResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func DeleteContainerHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vmid, err := strconv.Atoi(r.PathValue("vmid"))
		if err != nil || vmid <= 0 {
			http.Error(w, "vmid path parameter must be a positive integer", http.StatusBadRequest)
			return
		}

		body := ProxmoxVMStartRequest{VMID: vmid}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		req := coredeploy.ProxmoxVMStartRequest{
			HostServerID:    body.HostServerID,
			ProxmoxSecretID: body.ProxmoxSecretID,
			Node:            body.Node,
			VMID:            vmid,
		}

		result, err := service.DeleteContainer(r.Context(), req)
		if err != nil {
			slog.Error("failed to delete proxmox container", slog.Int("vmid", vmid), slog.String("error", err.Error()))
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
		body := ProxmoxLXCRequest{}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		req := coredeploy.ProxmoxLXCRequest{
			HostServerID:    body.HostServerID,
			ProxmoxSecretID: body.ProxmoxSecretID,
			Node:            body.Node,
			VMID:            body.VMID,
			Hostname:        body.Hostname,
			Password:        body.Password,
			OSTemplate:      body.OSTemplate,
			SshPublicKeys:   body.SshPublicKeys,
			Storage:         body.Storage,
			RootFSSize:      body.RootFSSize,
			Memory:          body.Memory,
			Swap:            body.Swap,
			Cores:           body.Cores,
			CPULimit:        body.CPULimit,
			CPUUnits:        body.CPUUnits,
			Net0:            body.Net0,
			Arch:            body.Arch,
			Cmode:           body.Cmode,
			Features:        body.Features,
			Nameserver:      body.Nameserver,
			SearchDomain:    body.SearchDomain,
			Description:     body.Description,
			Unprivileged:    body.Unprivileged,
			Start:           body.Start,
			Console:         body.Console,
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
		body := ProxmoxVMCreateRequest{}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		req := coredeploy.ProxmoxVMCreateRequest{
			HostServerID:      body.HostServerID,
			ProxmoxSecretID:   body.ProxmoxSecretID,
			Node:              body.Node,
			VMID:              body.VMID,
			TemplateVMID:      body.TemplateVMID,
			Name:              body.Name,
			MemoryMB:          body.MemoryMB,
			Sockets:           body.Sockets,
			Cores:             body.Cores,
			Description:       body.Description,
			Storage:           body.Storage,
			FullClone:         body.FullClone,
			Start:             body.Start,
			CIUser:            body.CIUser,
			CIPassword:        body.CIPassword,
			SshPublicKeys:     body.SshPublicKeys,
			IPConfig0:         body.IPConfig0,
			Nameserver:        body.Nameserver,
			SearchDomain:      body.SearchDomain,
			CISnippetsStorage: body.CISnippetsStorage,
			CICustomScript:    body.CICustomScript,
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
		body := ProxmoxVMTemplateRequest{}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		req := coredeploy.ProxmoxVMTemplateRequest{
			HostServerID:     body.HostServerID,
			ProxmoxSecretID:  body.ProxmoxSecretID,
			Node:             body.Node,
			VMID:             body.VMID,
			Name:             body.Name,
			ImageURL:         body.ImageURL,
			Storage:          body.Storage,
			CloudInitStorage: body.CloudInitStorage,
			MemoryMB:         body.MemoryMB,
			Sockets:          body.Sockets,
			Cores:            body.Cores,
			Description:      body.Description,
			Net0:             body.Net0,
			SCSIHW:           body.SCSIHW,
			DiskBus:          body.DiskBus,
			BootOrder:        body.BootOrder,
			Agent:            body.Agent,
			SerialConsole:    body.SerialConsole,
			CleanupImage:     body.CleanupImage,
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

// swagger:route POST /api/v1/proxmox/pve-user Proxmox CreateProxmoxPVEUser
// Create a Proxmox user over SSH on a Proxmox node.
// responses:
//
//	200: ProxmoxPVEUserCreateResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func CreatePVEUserHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := ProxmoxPVEUserCreateRequest{}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		req := coreproxmox.CreatePVEUserRequest{
			HostServerID: body.HostServerID,
			Node:         body.Node,
			Username:     body.Username,
			Realm:        body.Realm,
			Comment:      body.Comment,
			Password:     body.Password,
			Force:        body.Force,
		}

		result, err := service.CreatePVEUser(r.Context(), req)
		if err != nil {
			slog.Error("failed to create proxmox user", slog.String("error", err.Error()))
			writeServiceError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route POST /api/v1/proxmox/api-token Proxmox CreateProxmoxAPIToken
// Create a Proxmox API token over SSH on a Proxmox node.
// responses:
//
//	200: ProxmoxAPITokenCreateResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func CreateAPITokenHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := ProxmoxAPITokenCreateRequest{}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}
		req := coreproxmox.CreateAPITokenRequest{
			HostServerID:      body.HostServerID,
			Node:              body.Node,
			UserID:            body.UserID,
			Username:          body.Username,
			Realm:             body.Realm,
			TokenID:           body.TokenID,
			Comment:           body.Comment,
			Role:              body.Role,
			ACLPath:           body.ACLPath,
			ExpirationDate:    body.ExpirationDate,
			DaysValid:         body.DaysValid,
			Privsep:           body.Privsep,
			Force:             body.Force,
			Yolo:              body.Yolo,
			Verify:            body.Verify,
			StoreAsUserSecret: body.StoreAsUserSecret,
		}

		result, err := service.CreateAPIToken(r.Context(), req)
		if err != nil {
			slog.Error("failed to create proxmox API token", slog.String("error", err.Error()))
			writeServiceError(w, err)
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
	if hostServerID := r.URL.Query().Get("host_server_id"); hostServerID != "" {
		id, err := uuid.Parse(hostServerID)
		if err != nil {
			http.Error(w, "host_server_id must be a UUID", http.StatusBadRequest)
			return req, false
		}
		req.HostServerID = &id
	}
	if proxmoxSecretID := r.URL.Query().Get("proxmox_secret_id"); proxmoxSecretID != "" {
		id, err := uuid.Parse(proxmoxSecretID)
		if err != nil {
			http.Error(w, "proxmox_secret_id must be a UUID", http.StatusBadRequest)
			return req, false
		}
		req.ProxmoxSecretID = &id
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
