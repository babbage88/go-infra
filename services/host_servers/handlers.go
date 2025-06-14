package host_servers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/api/authapi"
	"github.com/google/uuid"
)

// swagger:route POST /host-servers/create host-servers createHostServer
// Create a new host server.
// responses:
//
//	200: HostServerResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func CreateHostServerHandler(provider HostServerProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateHostServerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Hostname == "" || req.IPAddress.IsValid() {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		server, err := provider.CreateHostServer(r.Context(), req)
		if err != nil {
			slog.Error("Failed to create host server", slog.String("error", err.Error()))
			http.Error(w, "Failed to create host server", http.StatusInternalServerError)
			return
		}

		resp := HostServerResponse{
			ID:               server.ID,
			Hostname:         server.Hostname,
			IPAddress:        server.IPAddress,
			IsContainerHost:  server.IsContainerHost,
			IsVmHost:         server.IsVmHost,
			IsVirtualMachine: server.IsVirtualMachine,
			IsDbHost:         server.IsDbHost,
			CreatedAt:        server.CreatedAt,
			LastModified:     server.LastModified,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
}

// swagger:route GET /host-servers/{ID} host-servers getHostServer
// Get a host server by ID.
// responses:
//
//	200: HostServerResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	404: description:Not Found
//	500: description:Internal Server Error
func GetHostServerHandler(provider HostServerProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlId := r.PathValue("ID")
		id, err := uuid.Parse(urlId)
		if err != nil {
			slog.Error("Failed to parse UUID", slog.String("error", err.Error()))
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		server, err := provider.GetHostServer(r.Context(), id)
		if err != nil {
			slog.Error("Failed to get host server", slog.String("error", err.Error()))
			http.Error(w, "Failed to get host server", http.StatusInternalServerError)
			return
		}

		if server == nil {
			http.Error(w, "Host server not found", http.StatusNotFound)
			return
		}

		resp := HostServerResponse{
			ID:               server.ID,
			Hostname:         server.Hostname,
			IPAddress:        server.IPAddress,
			IsContainerHost:  server.IsContainerHost,
			IsVmHost:         server.IsVmHost,
			IsVirtualMachine: server.IsVirtualMachine,
			IsDbHost:         server.IsDbHost,
			CreatedAt:        server.CreatedAt,
			LastModified:     server.LastModified,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
}

// swagger:route GET /host-servers host-servers getAllHostServers
// Get all host servers.
// responses:
//
//	200: HostServersResponse
//	401: description:Unauthorized
//	500: description:Internal Server Error
func GetAllHostServersHandler(provider HostServerProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		servers, err := provider.GetAllHostServers(r.Context())
		if err != nil {
			slog.Error("Failed to get all host servers", slog.String("error", err.Error()))
			http.Error(w, "Failed to get all host servers", http.StatusInternalServerError)
			return
		}

		resp := make(HostServersResponse, len(servers))
		for i, server := range servers {
			resp[i] = HostServerResponse{
				ID:               server.ID,
				Hostname:         server.Hostname,
				IPAddress:        server.IPAddress,
				IsContainerHost:  server.IsContainerHost,
				IsVmHost:         server.IsVmHost,
				IsVirtualMachine: server.IsVirtualMachine,
				IsDbHost:         server.IsDbHost,
				CreatedAt:        server.CreatedAt,
				LastModified:     server.LastModified,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
}

// swagger:route PUT /host-servers/{ID} host-servers updateHostServer
// Update a host server.
// responses:
//
//	200: HostServerResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	404: description:Not Found
//	500: description:Internal Server Error
func UpdateHostServerHandler(provider HostServerProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlId := r.PathValue("ID")
		id, err := uuid.Parse(urlId)
		if err != nil {
			slog.Error("Failed to parse UUID", slog.String("error", err.Error()))
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		var req UpdateHostServerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		server, err := provider.UpdateHostServer(r.Context(), id, req)
		if err != nil {
			slog.Error("Failed to update host server", slog.String("error", err.Error()))
			http.Error(w, "Failed to update host server", http.StatusInternalServerError)
			return
		}

		if server == nil {
			http.Error(w, "Host server not found", http.StatusNotFound)
			return
		}

		resp := HostServerResponse{
			ID:               server.ID,
			Hostname:         server.Hostname,
			IPAddress:        server.IPAddress,
			IsContainerHost:  server.IsContainerHost,
			IsVmHost:         server.IsVmHost,
			IsVirtualMachine: server.IsVirtualMachine,
			IsDbHost:         server.IsDbHost,
			CreatedAt:        server.CreatedAt,
			LastModified:     server.LastModified,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
}

// swagger:route DELETE /host-servers/{ID} host-servers deleteHostServer
// Delete a host server.
// responses:
//
//	200: description:Host server deleted successfully
//	400: description:Invalid request
//	401: description:Unauthorized
//	404: description:Not Found
//	500: description:Internal Server Error
func DeleteHostServerHandler(provider HostServerProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlId := r.PathValue("ID")
		id, err := uuid.Parse(urlId)
		if err != nil {
			slog.Error("Failed to parse UUID", slog.String("error", err.Error()))
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		err = provider.DeleteHostServer(r.Context(), id)
		if err != nil {
			slog.Error("Failed to delete host server", slog.String("error", err.Error()))
			http.Error(w, "Failed to delete host server", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
}
