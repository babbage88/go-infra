package user_applications

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// swagger:parameters createUserApplication
type CreateUserApplicationRequestWrapper struct {
	// in:body
	Body CreateUserApplicationRequest `json:"body"`
}

// swagger:parameters discoverUserApplication
type DiscoverUserApplicationRequestWrapper struct {
	// in:body
	Body DiscoverUserApplicationRequest `json:"body"`
}

// swagger:parameters updateUserApplication
type UpdateUserApplicationRequestWrapper struct {
	// In: path
	// Required: true
	ID string `json:"ID"`
	// in:body
	Body UpdateUserApplicationRequest `json:"body"`
}

// swagger:parameters getUserApplicationById
type GetUserApplicationByIdRequest struct {
	// In: path
	// Required: true
	ID string `json:"ID"`
}

// swagger:parameters getUserApplicationByName
type GetUserApplicationByNameRequest struct {
	// In: path
	// Required: true
	Name string `json:"name"`
}

// swagger:parameters deleteUserApplicationById
type DeleteUserApplicationByIdRequest struct {
	// In: path
	// Required: true
	ID string `json:"ID"`
}

// swagger:parameters deleteUserApplicationByName
type DeleteUserApplicationByNameRequest struct {
	// In: path
	// Required: true
	Name string `json:"name"`
}

// swagger:response UserApplicationResponse
type UserApplicationResponseWrapper struct {
	// in:body
	Body UserApplicationDao `json:"body"`
}

// swagger:response UserApplicationsResponse
type UserApplicationsResponseWrapper struct {
	// in:body
	Body []UserApplicationDao `json:"body"`
}

// swagger:response DiscoverUserApplicationResponse
type DiscoverUserApplicationResponseWrapper struct {
	// in:body
	Body DiscoverUserApplicationResponse `json:"body"`
}

// swagger:route POST /user-applications user-applications createUserApplication
// Register a deployable user application and its infrastructure dependencies.
//
// security:
// - bearer:
// responses:
// 201: UserApplicationResponse
// 400: description:Bad request - invalid input data
// 500: description:Internal server error
func CreateUserApplicationHandler(service UserApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateUserApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.Name == "" || req.RepositoryUrl == "" || req.DeployKind == "" {
			http.Error(w, "name, repositoryUrl, and deployKind are required", http.StatusBadRequest)
			return
		}

		app, err := service.CreateUserApplication(req)
		if err != nil {
			slog.Error("failed to create user application", slog.String("name", req.Name), slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Failed to create user application: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(app)
	}
}

// swagger:route POST /user-applications/discover user-applications discoverUserApplication
// Discover deployable application manifests from a source repository.
//
// security:
// - bearer:
// responses:
// 200: DiscoverUserApplicationResponse
// 400: description:Bad request - invalid input data
// 500: description:Internal server error
func DiscoverUserApplicationHandler(service UserApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DiscoverUserApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.RepositoryUrl == "" {
			http.Error(w, "repositoryUrl is required", http.StatusBadRequest)
			return
		}

		result, err := service.DiscoverUserApplication(req)
		if err != nil {
			slog.Error("failed to discover user application", slog.String("repositoryUrl", req.RepositoryUrl), slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Failed to discover user application: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}

// swagger:route GET /user-applications user-applications getAllUserApplications
// Get all registered user applications.
//
// security:
// - bearer:
// responses:
// 200: UserApplicationsResponse
// 500: description:Internal server error
func GetAllUserApplicationsHandler(service UserApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apps, err := service.GetAllUserApplications()
		if err != nil {
			slog.Error("failed to get all user applications", slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Failed to get user applications: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(apps)
	}
}

// swagger:route GET /user-applications/{ID} user-applications getUserApplicationById
// Get a registered user application by ID.
//
// security:
// - bearer:
// responses:
// 200: UserApplicationResponse
// 400: description:Bad request - invalid UUID format
// 404: description:User application not found
func GetUserApplicationByIdHandler(service UserApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("ID"))
		if err != nil {
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		app, err := service.GetUserApplicationById(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("User application not found: %v", err), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(app)
	}
}

// swagger:route GET /user-applications/by-name/{name} user-applications getUserApplicationByName
// Get a registered user application by name.
//
// security:
// - bearer:
// responses:
// 200: UserApplicationResponse
// 404: description:User application not found
func GetUserApplicationByNameHandler(service UserApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		app, err := service.GetUserApplicationByName(name)
		if err != nil {
			http.Error(w, fmt.Sprintf("User application not found: %v", err), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(app)
	}
}

// swagger:route PUT /user-applications/{ID} user-applications updateUserApplication
// Update a registered user application.
//
// security:
// - bearer:
// responses:
// 200: UserApplicationResponse
// 400: description:Bad request - invalid UUID format or request body
// 404: description:User application not found
func UpdateUserApplicationHandler(service UserApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("ID"))
		if err != nil {
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		var req UpdateUserApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		app, err := service.UpdateUserApplication(id, req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update user application: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(app)
	}
}

// swagger:route DELETE /user-applications/{ID} user-applications deleteUserApplicationById
// Delete a registered user application by ID.
//
// security:
// - bearer:
// responses:
// 204: description:User application deleted
// 400: description:Bad request - invalid UUID format
func DeleteUserApplicationByIdHandler(service UserApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("ID"))
		if err != nil {
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}
		if err := service.DeleteUserApplicationById(id); err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete user application: %v", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// swagger:route DELETE /user-applications/by-name/{name} user-applications deleteUserApplicationByName
// Delete a registered user application by name.
//
// security:
// - bearer:
// responses:
// 204: description:User application deleted
func DeleteUserApplicationByNameHandler(service UserApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		if name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		if err := service.DeleteUserApplicationByName(name); err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete user application: %v", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
