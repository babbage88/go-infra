package ssh_key_provider

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/api/authapi"
	"github.com/google/uuid"
)

// swagger:route POST /ssh-keys/create ssh-keys createSshKey
// Create a new SSH key.
// responses:
//
//	200: CreateSshKeyResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func CreateSshKeyHandler(provider SshKeySecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateSshKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Name == "" || req.PublicKey == "" || req.PrivateKey == "" || req.KeyType == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Get user ID from context
		userID, err := authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			slog.Error("Failed to get user ID from context", slog.String("error", err.Error()))
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Create the SSH key request
		sshKeyReq := &NewSshKeyRequest{
			UserID:      userID,
			Name:        req.Name,
			Description: req.Description,
			PublicKey:   req.PublicKey,
			PrivateKey:  req.PrivateKey,
			KeyType:     req.KeyType,
		}
		slog.Info("User ID", slog.String("user_id", userID.String()))
		slog.Info("Host server ID", slog.String("host_server_id", req.HostServerId.String()))
		slog.Info("Name", slog.String("name", req.Name))
		slog.Info("Description", slog.String("description", req.Description))
		slog.Info("Public key", slog.String("public_key", req.PublicKey))
		slog.Info("Private key", slog.String("private_key", req.PrivateKey))
		slog.Info("Key type", slog.String("key_type", req.KeyType))

		// Add host server ID if provided
		if req.HostServerId != nil {
			slog.Info("Host server ID", slog.String("host_server_id", req.HostServerId.String()))
			sshKeyReq.HostServerId = *req.HostServerId
		}

		// Create the SSH key
		result := provider.CreateSshKey(sshKeyReq)
		slog.Info("Created SSH key", slog.String("result", fmt.Sprintf("%+v", result)))
		if result.Error != nil {
			slog.Error("Failed to create SSH key", slog.String("error", result.Error.Error()))
			http.Error(w, "Failed to create SSH key", http.StatusInternalServerError)
			return
		}

		// Prepare response
		resp := CreateSshKeyResponse{
			SshKeyId:        result.SshKeyId,
			PrivKeySecretId: result.PrivKeySecretId,
			UserId:          result.UserId,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
}

// swagger:route DELETE /ssh-keys/{id} ssh-keys deleteSshKey
// Delete an SSH key and its associated secret.
//
// responses:
//
//	200: DeleteSshKeyResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	404: description:Not Found
//	500: description:Internal Server Error
func DeleteSshKeyHandler(provider SshKeySecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "Missing SSH key ID", http.StatusBadRequest)
			return
		}

		sshKeyId, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid SSH key ID", http.StatusBadRequest)
			return
		}

		// Get user ID from context (for future RBAC, not used here)
		_, err = authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		err = provider.DeleteSShKeyAndSecret(sshKeyId)
		if err != nil {
			slog.Error("Failed to delete SSH key and secret", slog.String("error", err.Error()))
			if err.Error() == "no rows in result set" {
				http.Error(w, "SSH key not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to delete SSH key and secret", http.StatusInternalServerError)
			return
		}

		resp := DeleteSshKeyResponse{
			Body: struct {
				Message string `json:"message"`
			}{Message: "SSH key and secret deleted successfully"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

// SSH Key Host Mapping CRUD Handlers

// swagger:route POST /ssh-key-host-mappings/create ssh-key-host-mappings createSshKeyHostMapping
// Create a new SSH key host mapping.
// responses:
//
//	200: CreateSshKeyHostMappingResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func CreateSshKeyHostMappingHandler(provider SshKeySecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateSshKeyHostMappingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.SshKeyID == uuid.Nil || req.HostServerID == uuid.Nil || req.UserID == uuid.Nil || req.HostserverUsername == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Create the SSH key host mapping
		result := provider.CreateSshKeyHostMapping(&req)
		if result.Error != nil {
			slog.Error("Failed to create SSH key host mapping", slog.String("error", result.Error.Error()))
			http.Error(w, "Failed to create SSH key host mapping", http.StatusInternalServerError)
			return
		}

		// Prepare response
		resp := CreateSshKeyHostMappingResponse{
			ID:                 result.ID,
			SshKeyID:           result.SshKeyID,
			HostServerID:       result.HostServerID,
			UserID:             result.UserID,
			HostserverUsername: result.HostserverUsername,
			CreatedAt:          result.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastModified:       result.LastModified.Format("2006-01-02T15:04:05Z07:00"),
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
}

// swagger:route GET /ssh-key-host-mappings/{id} ssh-key-host-mappings getSshKeyHostMappingById
// Get an SSH key host mapping by ID.
// responses:
//
//	200: GetSshKeyHostMappingByIdResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	404: description:Not Found
//	500: description:Internal Server Error
func GetSshKeyHostMappingByIdHandler(provider SshKeySecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "Missing mapping ID", http.StatusBadRequest)
			return
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid mapping ID", http.StatusBadRequest)
			return
		}

		// Get user ID from context (for future RBAC, not used here)
		_, err = authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		result, err := provider.GetSshKeyHostMappingById(id)
		if err != nil {
			slog.Error("Failed to get SSH key host mapping", slog.String("error", err.Error()))
			if err.Error() == "no rows in result set" {
				http.Error(w, "SSH key host mapping not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to get SSH key host mapping", http.StatusInternalServerError)
			return
		}

		// Prepare response
		resp := CreateSshKeyHostMappingResponse{
			ID:                 result.ID,
			SshKeyID:           result.SshKeyID,
			HostServerID:       result.HostServerID,
			UserID:             result.UserID,
			HostserverUsername: result.HostserverUsername,
			CreatedAt:          result.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastModified:       result.LastModified.Format("2006-01-02T15:04:05Z07:00"),
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
}

// swagger:route GET /ssh-key-host-mappings/user/{userId} ssh-key-host-mappings getSshKeyHostMappingsByUserId
// Get all SSH key host mappings for a user.
// responses:
//
//	200: GetSshKeyHostMappingsByUserIdResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func GetSshKeyHostMappingsByUserIdHandler(provider SshKeySecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIdStr := r.PathValue("userId")
		if userIdStr == "" {
			http.Error(w, "Missing user ID", http.StatusBadRequest)
			return
		}

		userId, err := uuid.Parse(userIdStr)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Get user ID from context (for future RBAC, not used here)
		_, err = authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		results, err := provider.GetSshKeyHostMappingsByUserId(userId)
		if err != nil {
			slog.Error("Failed to get SSH key host mappings by user ID", slog.String("error", err.Error()))
			http.Error(w, "Failed to get SSH key host mappings", http.StatusInternalServerError)
			return
		}

		// Prepare response
		resp := make([]CreateSshKeyHostMappingResponse, 0, len(results))
		for _, result := range results {
			resp = append(resp, CreateSshKeyHostMappingResponse{
				ID:                 result.ID,
				SshKeyID:           result.SshKeyID,
				HostServerID:       result.HostServerID,
				UserID:             result.UserID,
				HostserverUsername: result.HostserverUsername,
				CreatedAt:          result.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				LastModified:       result.LastModified.Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
}

// swagger:route GET /ssh-key-host-mappings/host/{hostId} ssh-key-host-mappings getSshKeyHostMappingsByHostId
// Get all SSH key host mappings for a host server.
// responses:
//
//	200: GetSshKeyHostMappingsByHostIdResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func GetSshKeyHostMappingsByHostIdHandler(provider SshKeySecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hostIdStr := r.PathValue("hostId")
		if hostIdStr == "" {
			http.Error(w, "Missing host ID", http.StatusBadRequest)
			return
		}

		hostId, err := uuid.Parse(hostIdStr)
		if err != nil {
			http.Error(w, "Invalid host ID", http.StatusBadRequest)
			return
		}

		// Get user ID from context (for future RBAC, not used here)
		_, err = authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		results, err := provider.GetSshKeyHostMappingsByHostId(hostId)
		if err != nil {
			slog.Error("Failed to get SSH key host mappings by host ID", slog.String("error", err.Error()))
			http.Error(w, "Failed to get SSH key host mappings", http.StatusInternalServerError)
			return
		}

		// Prepare response
		resp := make([]CreateSshKeyHostMappingResponse, 0, len(results))
		for _, result := range results {
			resp = append(resp, CreateSshKeyHostMappingResponse{
				ID:                 result.ID,
				SshKeyID:           result.SshKeyID,
				HostServerID:       result.HostServerID,
				UserID:             result.UserID,
				HostserverUsername: result.HostserverUsername,
				CreatedAt:          result.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				LastModified:       result.LastModified.Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
}

// swagger:route GET /ssh-key-host-mappings/key/{keyId} ssh-key-host-mappings getSshKeyHostMappingsByKeyId
// Get all SSH key host mappings for an SSH key.
// responses:
//
//	200: GetSshKeyHostMappingsByKeyIdResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func GetSshKeyHostMappingsByKeyIdHandler(provider SshKeySecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keyIdStr := r.PathValue("keyId")
		if keyIdStr == "" {
			http.Error(w, "Missing key ID", http.StatusBadRequest)
			return
		}

		keyId, err := uuid.Parse(keyIdStr)
		if err != nil {
			http.Error(w, "Invalid key ID", http.StatusBadRequest)
			return
		}

		// Get user ID from context (for future RBAC, not used here)
		_, err = authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		results, err := provider.GetSshKeyHostMappingsByKeyId(keyId)
		if err != nil {
			slog.Error("Failed to get SSH key host mappings by key ID", slog.String("error", err.Error()))
			http.Error(w, "Failed to get SSH key host mappings", http.StatusInternalServerError)
			return
		}

		// Prepare response
		resp := make([]CreateSshKeyHostMappingResponse, 0, len(results))
		for _, result := range results {
			resp = append(resp, CreateSshKeyHostMappingResponse{
				ID:                 result.ID,
				SshKeyID:           result.SshKeyID,
				HostServerID:       result.HostServerID,
				UserID:             result.UserID,
				HostserverUsername: result.HostserverUsername,
				CreatedAt:          result.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				LastModified:       result.LastModified.Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
}

// swagger:route PUT /ssh-key-host-mappings/{id} ssh-key-host-mappings updateSshKeyHostMapping
// Update an SSH key host mapping.
// responses:
//
//	200: UpdateSshKeyHostMappingResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	404: description:Not Found
//	500: description:Internal Server Error
func UpdateSshKeyHostMappingHandler(provider SshKeySecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "Missing mapping ID", http.StatusBadRequest)
			return
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid mapping ID", http.StatusBadRequest)
			return
		}

		var req UpdateSshKeyHostMappingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Failed to decode request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.HostserverUsername == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Set the ID from the path
		req.ID = id

		// Get user ID from context (for future RBAC, not used here)
		_, err = authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		result := provider.UpdateSshKeyHostMapping(&req)
		if result.Error != nil {
			slog.Error("Failed to update SSH key host mapping", slog.String("error", result.Error.Error()))
			if result.Error.Error() == "no rows in result set" {
				http.Error(w, "SSH key host mapping not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to update SSH key host mapping", http.StatusInternalServerError)
			return
		}

		// Prepare response
		resp := CreateSshKeyHostMappingResponse{
			ID:                 result.ID,
			SshKeyID:           result.SshKeyID,
			HostServerID:       result.HostServerID,
			UserID:             result.UserID,
			HostserverUsername: result.HostserverUsername,
			CreatedAt:          result.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			LastModified:       result.LastModified.Format("2006-01-02T15:04:05Z07:00"),
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			slog.Error("Failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
}

// swagger:route DELETE /ssh-key-host-mappings/{id} ssh-key-host-mappings deleteSshKeyHostMapping
// Delete an SSH key host mapping.
// responses:
//
//	200: DeleteSshKeyHostMappingResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	404: description:Not Found
//	500: description:Internal Server Error
func DeleteSshKeyHostMappingHandler(provider SshKeySecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "Missing mapping ID", http.StatusBadRequest)
			return
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "Invalid mapping ID", http.StatusBadRequest)
			return
		}

		// Get user ID from context (for future RBAC, not used here)
		_, err = authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		err = provider.DeleteSshKeyHostMapping(id)
		if err != nil {
			slog.Error("Failed to delete SSH key host mapping", slog.String("error", err.Error()))
			if err.Error() == "no rows in result set" {
				http.Error(w, "SSH key host mapping not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to delete SSH key host mapping", http.StatusInternalServerError)
			return
		}

		resp := DeleteSshKeyHostMappingResponse{
			Message: "SSH key host mapping deleted successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

// SshKeyHostMappingByIDHandler handles GET, PUT, and DELETE operations for SSH key host mappings by ID
func SshKeyHostMappingByIDHandler(provider SshKeySecretProvider, authService authapi.AuthService) http.Handler {
	return authapi.AuthMiddlewareRequirePermission(authService, "ManageSshKeys", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetSshKeyHostMappingByIdHandler(provider).ServeHTTP(w, r)
		case http.MethodPut:
			UpdateSshKeyHostMappingHandler(provider).ServeHTTP(w, r)
		case http.MethodDelete:
			DeleteSshKeyHostMappingHandler(provider).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
}
