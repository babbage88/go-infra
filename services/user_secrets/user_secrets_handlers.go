package user_secrets

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/api/authapi"
	"github.com/google/uuid"
)

// swagger:route POST /secrets/create secrets createUserSecret
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

		err = provider.StoreSecret(req.Secret, userID, req.ApplicationID, req.Expiration)
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

		if secret.ExternalAuthToken.UserID != userID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		resp := RetrievedSecretResponse{}
		resp.Body.ID = secret.ExternalAuthToken.Id
		resp.Body.UserID = secret.ExternalAuthToken.UserID
		resp.Body.ExternalApplication = secret.ExternalAuthToken.ExternalApplicationId
		resp.Body.Expiration = secret.ExternalAuthToken.Expiration
		resp.Body.Secret = string(secret.ExternalAuthToken.Token)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp.Body)
	}))
}

// swagger:route GET /user/secrets/{USERID} secrets GetUserSecretEntries
// Retrieve a user secret by USERID.
// responses:
//
//	200: GetUserSecretEntriesResponse
//	401: description:Unauthorized
//	403: description:Forbidden
//	404: description:Not Found

func GetUserSecretEntriesByIdHandler(provider UserSecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlId := r.PathValue("USERID")
		urlUserId, err := uuid.Parse(urlId)
		if err != nil {
			slog.Error("Error parsing UUID from url string")
			http.Error(w, "Invalid USERID", http.StatusBadRequest)
			return
		}

		userID, err := authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if urlUserId != userID {
			http.Error(w, "unauthorized to access secrets for requested user id", http.StatusUnauthorized)

		}

		secrets, err := provider.GetUserSecretEntries(urlUserId)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if secrets == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
		}

		resp := GetUserSecretEntriesResponseWrapper{
			Body: secrets,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp.Body)

	}))
}

// swagger:route GET /user/{APPID}/secrets/{USERID} secrets GetUserSecretEntriesByAppId
// Retrieve a user secret by USERID.
// responses:
//
//	200: GetUserSecretEntriesResponse
//	401: description:Unauthorized
//	403: description:Forbidden
//	404: description:Not Found

func GetUserSecretEntriesByAppIdHandler(provider UserSecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlappId := r.PathValue("APPID")
		urlId := r.PathValue("USERID")
		urlUserId, err := uuid.Parse(urlId)
		if err != nil {
			slog.Error("Error parsing USER UUID from url string")
			http.Error(w, "Invalid USERID", http.StatusBadRequest)
		}

		urlAppId, err := uuid.Parse(urlappId)
		if err != nil {
			slog.Error("Error parsing APP UUID from url string")
			http.Error(w, "Invalid USERID", http.StatusBadRequest)
		}

		userID, err := authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if urlUserId != userID {
			http.Error(w, "unauthorized to access secrets for requested user id", http.StatusUnauthorized)

		}

		secrets, err := provider.GetUserSecretEntriesByAppId(urlUserId, urlAppId)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if secrets == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
		}

		resp := GetUserSecretEntriesResponseWrapper{
			Body: secrets,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp.Body)

	}))
}

// swagger:route DELETE /secrets/delete/{ID} secrets deleteUserSecretByID
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

		if secret.ExternalAuthToken.UserID != userID {
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

// swagger:route GET /user/{APPNAME}/secrets/{USERID} secrets GetUserSecretEntriesByAppName
// Retrieve a user secret by USERID and application name.
// responses:
//
//	200: GetUserSecretEntriesResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	403: description:Forbidden
//	404: description:Not Found
//	500: description:Internal Server Error
func GetUserSecretEntriesByAppNameHandler(provider UserSecretProvider) http.Handler {
	return authapi.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appName := r.PathValue("APPNAME")
		urlId := r.PathValue("USERID")
		urlUserId, err := uuid.Parse(urlId)
		if err != nil {
			slog.Error("Error parsing USER UUID from url string")
			http.Error(w, "Invalid USERID", http.StatusBadRequest)
			return
		}

		userID, err := authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if urlUserId != userID {
			http.Error(w, "unauthorized to access secrets for requested user id", http.StatusUnauthorized)
			return
		}

		secrets, err := provider.GetUserSecretEntriesByAppName(urlUserId, appName)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if secrets == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		resp := GetUserSecretEntriesResponseWrapper{
			Body: secrets,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp.Body)
	}))
}
