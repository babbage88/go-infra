package user_secrets

import (
	"encoding/json"
	"net/http"

	"github.com/babbage88/go-infra/webapi/authapi"
	"github.com/google/uuid"
)

// swagger:route POST /user/secrets/create secrets createUserSecret
// Create a new external application secret.
// responses:
//
//	200: description:Secret stored successfully
//	400: description:Invalid request
//	401: description:Unauthorized
func CreateSecretHandler(provider UserSecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CreateSecretRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		userID, err := authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		err = provider.StoreSecret(req.Secret, userID, req.ApplicationID)
		if err != nil {
			http.Error(w, "Failed to store secret", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
}

// swagger:route GET /secrets/{ID} secrets getUserSecretByID
// Retrieve a user secret by ID.
// responses:
//
//	200: RetrievedSecretResponse
//	401: description:Unauthorized
//	403: description:Forbidden
//	404: description:Not Found
func GetSecretHandler(provider UserSecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlId := r.PathValue("ID")
		secretId, err := uuid.Parse(urlId)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		userID, err := authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		secret, err := provider.RetrieveSecret(secretId)
		if err != nil || secret == nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		if secret.Metadata.UserID != userID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		resp := RetrievedSecretResponse{}
		resp.Body.ID = secret.Metadata.Id
		resp.Body.UserID = secret.Metadata.UserID
		resp.Body.ExternalApplication = secret.Metadata.ExternalApplicationId
		resp.Body.Expiration = secret.Metadata.Expiration
		resp.Body.Token = string(secret.Metadata.Token)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp.Body)
	}))
}

// swagger:route DELETE /secrets/{id} secrets deleteUserSecretByID
// Delete a user secret by ID.
// responses:
//
//	200: description:Secret deleted successfully
//	401: description:Unauthorized
//	403: description:Forbidden
//	404: description:Not Found
func DeleteSecretHandler(provider UserSecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlId := r.PathValue("ID")
		secretId, err := uuid.Parse(urlId)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		userID, err := authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		secret, err := provider.RetrieveSecret(secretId)
		if err != nil || secret == nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		if secret.Metadata.UserID != userID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if err := provider.DeleteSecret(secretId); err != nil {
			http.Error(w, "Failed to delete secret", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
}
