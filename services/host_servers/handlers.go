package host_servers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// swagger:route POST /host-servers/create host-servers CreateHostServer
// Create a new host server.
// responses:
//
//	200: HostServerResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func CreateHostServerHandler(provider HostServerProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateHostServerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Hostname == "" || !req.IPAddress.IsValid() {
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
			ID:                  server.ID,
			Hostname:            server.Hostname,
			IPAddress:           server.IPAddress,
			Username:            server.Username,
			SSHKeyID:            server.SSHKeyID,
			SudoPasswordTokenID: server.SudoPasswordSecretID,
			HostServerTypes:     server.HostServerTypes,
			PlatformTypes:       server.PlatformTypes,
			CreatedAt:           server.CreatedAt,
			LastModified:        server.LastModified,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// swagger:route GET /host-servers/{ID} host-servers GetHostServer
// Get a host server by ID.
// responses:
//
//	200: HostServerResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	404: description:Not Found
//	500: description:Internal Server Error
func GetHostServerHandler(provider HostServerProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			ID:                  server.ID,
			Hostname:            server.Hostname,
			IPAddress:           server.IPAddress,
			Username:            server.Username,
			SSHKeyID:            server.SSHKeyID,
			SudoPasswordTokenID: server.SudoPasswordSecretID,
			HostServerTypes:     server.HostServerTypes,
			PlatformTypes:       server.PlatformTypes,
			CreatedAt:           server.CreatedAt,
			LastModified:        server.LastModified,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// swagger:route GET /host-servers host-servers GetAllHostServers
// Get all host servers.
// responses:
//
//	200: HostServersResponse
//	401: description:Unauthorized
//	500: description:Internal Server Error
func GetAllHostServersHandler(provider HostServerProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		servers, err := provider.GetAllHostServers(r.Context())
		if err != nil {
			slog.Error("Failed to get all host servers", slog.String("error", err.Error()))
			http.Error(w, "Failed to get all host servers", http.StatusInternalServerError)
			return
		}

		respSlice := make(HostServersResponse, len(servers))
		for i, server := range servers {
			respSlice[i] = HostServerResponse{
				ID:                  server.ID,
				Hostname:            server.Hostname,
				IPAddress:           server.IPAddress,
				Username:            server.Username,
				SSHKeyID:            server.SSHKeyID,
				SudoPasswordTokenID: server.SudoPasswordSecretID,
				HostServerTypes:     server.HostServerTypes,
				PlatformTypes:       server.PlatformTypes,
				CreatedAt:           server.CreatedAt,
				LastModified:        server.LastModified,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respSlice); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// swagger:route PUT /host-servers/{ID} host-servers UpdateHostServer
// Update a host server.
// responses:
//
//	200: HostServerResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	404: description:Not Found
//	500: description:Internal Server Error
func UpdateHostServerHandler(provider HostServerProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			ID:                  server.ID,
			Hostname:            server.Hostname,
			IPAddress:           server.IPAddress,
			Username:            server.Username,
			SSHKeyID:            server.SSHKeyID,
			SudoPasswordTokenID: server.SudoPasswordSecretID,
			HostServerTypes:     server.HostServerTypes,
			PlatformTypes:       server.PlatformTypes,
			CreatedAt:           server.CreatedAt,
			LastModified:        server.LastModified,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// swagger:route DELETE /host-servers/{ID} host-servers DeleteHostServer
// Delete a host server.
// responses:
//
//	200: description:Host server deleted successfully
//	400: description:Invalid request
//	401: description:Unauthorized
//	404: description:Not Found
//	500: description:Internal Server Error
func DeleteHostServerHandler(provider HostServerProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

// swagger:route GET /host-server-types host-servers GetAllHostServerTypes
// Get all available host server types.
// responses:
//
//	200: GetAllHostServerTypesResponse
//	401: description:Unauthorized
//	500: description:Internal Server Error
func GetAllHostServerTypesHandler(provider HostServerProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hostServerTypes, err := provider.GetAllHostServerTypes(r.Context())
		if err != nil {
			slog.Error("Failed to get all host server types", slog.String("error", err.Error()))
			http.Error(w, "Failed to get all host server types", http.StatusInternalServerError)
			return
		}

		respSlice := make([]HostServerType, len(hostServerTypes))
		for i, hostServerType := range hostServerTypes {
			respSlice[i] = HostServerType{
				ID:           hostServerType.ID,
				Name:         hostServerType.Name,
				LastModified: hostServerType.LastModified,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respSlice); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// swagger:route GET /platform-types host-servers GetAllPlatformTypes
// Get all available platform types.
// responses:
//
//	200: GetAllPlatformTypesResponse
//	401: description:Unauthorized
//	500: description:Internal Server Error
func GetAllPlatformTypesHandler(provider HostServerProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		platformTypes, err := provider.GetAllPlatformTypes(r.Context())
		if err != nil {
			slog.Error("Failed to get all platform types", slog.String("error", err.Error()))
			http.Error(w, "Failed to get all platform types", http.StatusInternalServerError)
			return
		}

		respSlice := make([]PlatformType, len(platformTypes))
		for i, platformType := range platformTypes {
			respSlice[i] = PlatformType{
				ID:           platformType.ID,
				Name:         platformType.Name,
				LastModified: platformType.LastModified,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respSlice); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// swagger:route POST /host-server-type-mappings host-servers CreateHostServerTypeMapping
// Create a mapping between a host server and a host server type.
// responses:
//
//	200: CreateHostServerTypeMappingResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func CreateHostServerTypeMappingHandler(provider HostServerProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateHostServerTypeMappingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.HostServerId == uuid.Nil || req.HostServerTypeId == uuid.Nil {
			http.Error(w, "hostServerId and hostServerTypeId are required", http.StatusBadRequest)
			return
		}
		if err := provider.CreateHostServerTypeMapping(r.Context(), req.HostServerId, req.HostServerTypeId); err != nil {
			http.Error(w, "Failed to create mapping: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Success bool `json:"success"`
		}{Success: true})
	}
}

// swagger:route POST /platform-type-mappings host-servers CreatePlatformTypeMapping
// Create a mapping between a host server, platform type, and host server type.
// responses:
//
//	200: CreatePlatformTypeMappingResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func CreatePlatformTypeMappingHandler(provider HostServerProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreatePlatformTypeMappingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.HostServerId == uuid.Nil || req.PlatformTypeId == uuid.Nil || req.HostServerTypeId == uuid.Nil {
			http.Error(w, "hostServerId, platformTypeId, and hostServerTypeId are required", http.StatusBadRequest)
			return
		}
		if err := provider.CreatePlatformTypeMapping(r.Context(), req.HostServerId, req.PlatformTypeId, req.HostServerTypeId); err != nil {
			http.Error(w, "Failed to create mapping: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Success bool `json:"success"`
		}{Success: true})
	}
}
