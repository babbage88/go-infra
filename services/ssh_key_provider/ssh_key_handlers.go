package ssh_key_provider

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/api/authapi"
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

		// Add host server ID if provided
		if req.HostServerId != nil {
			sshKeyReq.HostServerId = *req.HostServerId
		}

		// Create the SSH key
		result := provider.CreateSshKey(sshKeyReq)
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
