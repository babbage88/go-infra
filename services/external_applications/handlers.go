package external_applications

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	authapi "github.com/babbage88/go-infra/api/authapi"
	"github.com/google/uuid"
)

// swagger:parameters createExternalApplication
type CreateExternalApplicationRequestWrapper struct {
	// in:body
	Body CreateExternalApplicationRequest `json:"body"`
}

// swagger:response CreateExternalApplicationResponse
type CreateExternalApplicationResponseWrapper struct {
	// in:body
	Body ExternalApplicationDao `json:"body"`
}

// swagger:parameters getExternalApplicationById
type GetExternalApplicationByIdRequest struct {
	// ID of the external application
	//
	// In: path
	// Required: true
	ID string `json:"ID"`
}

// swagger:response GetExternalApplicationByIdResponse
type GetExternalApplicationByIdResponseWrapper struct {
	// in:body
	Body ExternalApplicationDao `json:"body"`
}

// swagger:parameters getExternalApplicationByName
type GetExternalApplicationByNameRequest struct {
	// Name of the external application
	//
	// In: path
	// Required: true
	Name string `json:"name"`
}

// swagger:response GetExternalApplicationByNameResponse
type GetExternalApplicationByNameResponseWrapper struct {
	// in:body
	Body ExternalApplicationDao `json:"body"`
}

// swagger:response GetAllExternalApplicationsResponse
type GetAllExternalApplicationsResponseWrapper struct {
	// in:body
	Body []ExternalApplicationDao `json:"body"`
}

// swagger:parameters updateExternalApplication
type UpdateExternalApplicationRequestWrapper struct {
	// ID of the external application
	//
	// In: path
	// Required: true
	ID string `json:"ID"`
	// in:body
	Body UpdateExternalApplicationRequest `json:"body"`
}

// swagger:response UpdateExternalApplicationResponse
type UpdateExternalApplicationResponseWrapper struct {
	// in:body
	Body ExternalApplicationDao `json:"body"`
}

// swagger:parameters deleteExternalApplicationById
type DeleteExternalApplicationByIdRequest struct {
	// ID of the external application
	//
	// In: path
	// Required: true
	ID string `json:"ID"`
}

// swagger:parameters deleteExternalApplicationByName
type DeleteExternalApplicationByNameRequest struct {
	// Name of the external application
	//
	// In: path
	// Required: true
	Name string `json:"name"`
}

// swagger:parameters getExternalApplicationIdByName
type GetExternalApplicationIdByNameRequest struct {
	// Name of the external application
	//
	// In: path
	// Required: true
	Name string `json:"name"`
}

// swagger:response GetExternalApplicationIdByNameResponse
type GetExternalApplicationIdByNameResponseWrapper struct {
	// in:body
	Body struct {
		ID uuid.UUID `json:"id"`
	} `json:"body"`
}

// swagger:parameters getExternalApplicationNameById
type GetExternalApplicationNameByIdRequest struct {
	// ID of the external application
	//
	// In: path
	// Required: true
	ID string `json:"ID"`
}

// swagger:response GetExternalApplicationNameByIdResponse
type GetExternalApplicationNameByIdResponseWrapper struct {
	// in:body
	Body struct {
		Name string `json:"name"`
	} `json:"body"`
}

// CreateExternalApplicationHandler handles POST requests to create a new external application
// swagger:operation POST /external-applications external-applications createExternalApplication
//
// # Create a new external application
//
// ---
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
//   - name: body
//     in: body
//     required: true
//     schema:
//     "$ref": "#/definitions/CreateExternalApplicationRequest"
//
// responses:
//
//	"201":
//	  description: External application created successfully
//	  schema:
//	    "$ref": "#/definitions/ExternalApplicationDao"
//	"400":
//	  description: Bad request - invalid input data
//	"409":
//	  description: Conflict - application with this name already exists
//	"500":
//	  description: Internal server error
func CreateExternalApplicationHandler(service ExternalApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateExternalApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Error decoding request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		app, err := service.CreateExternalApplication(req)
		if err != nil {
			slog.Error("Error creating external application",
				slog.String("name", req.Name),
				slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Failed to create external application: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(app)
	}
}

// GetExternalApplicationByIdHandler handles GET requests to retrieve an external application by ID
// swagger:operation GET /external-applications/{ID} external-applications getExternalApplicationById
//
// # Get an external application by ID
//
// ---
// produces:
// - application/json
// parameters:
//   - name: ID
//     in: path
//     required: true
//     type: string
//     format: uuid
//
// responses:
//
//	"200":
//	  description: External application retrieved successfully
//	  schema:
//	    "$ref": "#/definitions/ExternalApplicationDao"
//	"404":
//	  description: External application not found
//	"400":
//	  description: Bad request - invalid UUID format
//	"500":
//	  description: Internal server error
func GetExternalApplicationByIdHandler(service ExternalApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("ID")

		id, err := uuid.Parse(idStr)
		if err != nil {
			slog.Error("Error parsing UUID", slog.String("id", idStr), slog.String("error", err.Error()))
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		app, err := service.GetExternalApplicationById(id)
		if err != nil {
			slog.Error("Error getting external application by ID",
				slog.String("id", id.String()),
				slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("External application not found: %v", err), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(app)
	}
}

// GetExternalApplicationByNameHandler handles GET requests to retrieve an external application by name
// swagger:operation GET /external-applications/by-name/{name} external-applications getExternalApplicationByName
//
// # Get an external application by name
//
// ---
// produces:
// - application/json
// parameters:
//   - name: name
//     in: path
//     required: true
//     type: string
//
// responses:
//
//	"200":
//	  description: External application retrieved successfully
//	  schema:
//	    "$ref": "#/definitions/ExternalApplicationDao"
//	"404":
//	  description: External application not found
//	"500":
//	  description: Internal server error
func GetExternalApplicationByNameHandler(service ExternalApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")

		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		app, err := service.GetExternalApplicationByName(name)
		if err != nil {
			slog.Error("Error getting external application by name",
				slog.String("name", name),
				slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("External application not found: %v", err), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(app)
	}
}

// GetAllExternalApplicationsHandler handles GET requests to retrieve all external applications
// swagger:operation GET /external-applications external-applications getAllExternalApplications
//
// # Get all external applications
//
// ---
// produces:
// - application/json
// responses:
//
//	"200":
//	  description: External applications retrieved successfully
//	  schema:
//	    type: array
//	    items:
//	      "$ref": "#/definitions/ExternalApplicationDao"
//	"500":
//	  description: Internal server error
func GetAllExternalApplicationsHandler(service ExternalApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apps, err := service.GetAllExternalApplications()
		if err != nil {
			slog.Error("Error getting all external applications", slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Failed to get external applications: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apps)
	}
}

// UpdateExternalApplicationHandler handles PUT requests to update an external application
// swagger:operation PUT /external-applications/{ID} external-applications updateExternalApplication
//
// # Update an external application
//
// ---
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
//   - name: ID
//     in: path
//     required: true
//     type: string
//     format: uuid
//   - name: body
//     in: body
//     required: true
//     schema:
//     "$ref": "#/definitions/UpdateExternalApplicationRequest"
//
// responses:
//
//	"200":
//	  description: External application updated successfully
//	  schema:
//	    "$ref": "#/definitions/ExternalApplicationDao"
//	"404":
//	  description: External application not found
//	"400":
//	  description: Bad request - invalid UUID format or request body
//	"500":
//	  description: Internal server error
func UpdateExternalApplicationHandler(service ExternalApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("ID")

		id, err := uuid.Parse(idStr)
		if err != nil {
			slog.Error("Error parsing UUID", slog.String("id", idStr), slog.String("error", err.Error()))
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		var req UpdateExternalApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("Error decoding request body", slog.String("error", err.Error()))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		app, err := service.UpdateExternalApplication(id, req)
		if err != nil {
			slog.Error("Error updating external application",
				slog.String("id", id.String()),
				slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Failed to update external application: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(app)
	}
}

// DeleteExternalApplicationByIdHandler handles DELETE requests to delete an external application by ID
// swagger:operation DELETE /external-applications/{ID} external-applications deleteExternalApplicationById
//
// # Delete an external application by ID
//
// ---
// parameters:
//   - name: ID
//     in: path
//     required: true
//     type: string
//     format: uuid
//
// responses:
//
//	"204":
//	  description: External application deleted successfully
//	"404":
//	  description: External application not found
//	"400":
//	  description: Bad request - invalid UUID format
//	"500":
//	  description: Internal server error
func DeleteExternalApplicationByIdHandler(service ExternalApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("ID")

		id, err := uuid.Parse(idStr)
		if err != nil {
			slog.Error("Error parsing UUID", slog.String("id", idStr), slog.String("error", err.Error()))
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		err = service.DeleteExternalApplicationById(id)
		if err != nil {
			slog.Error("Error deleting external application by ID",
				slog.String("id", id.String()),
				slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Failed to delete external application: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// DeleteExternalApplicationByNameHandler handles DELETE requests to delete an external application by name
// swagger:operation DELETE /external-applications/by-name/{name} external-applications deleteExternalApplicationByName
//
// # Delete an external application by name
//
// ---
// parameters:
//   - name: name
//     in: path
//     required: true
//     type: string
//
// responses:
//
//	"204":
//	  description: External application deleted successfully
//	"404":
//	  description: External application not found
//	"500":
//	  description: Internal server error
func DeleteExternalApplicationByNameHandler(service ExternalApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")

		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		err := service.DeleteExternalApplicationByName(name)
		if err != nil {
			slog.Error("Error deleting external application by name",
				slog.String("name", name),
				slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("Failed to delete external application: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// GetExternalApplicationIdByNameHandler handles GET requests to retrieve an external application ID by name
// swagger:operation GET /external-applications/id/{name} external-applications getExternalApplicationIdByName
//
// # Get an external application ID by name
//
// ---
// produces:
// - application/json
// parameters:
//   - name: name
//     in: path
//     required: true
//     type: string
//
// responses:
//
//	"200":
//	  description: External application ID retrieved successfully
//	  schema:
//	    type: object
//	    properties:
//	      id:
//	        type: string
//	        format: uuid
//	"404":
//	  description: External application not found
//	"500":
//	  description: Internal server error
func GetExternalApplicationIdByNameHandler(service ExternalApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")

		if name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		id, err := service.GetExternalApplicationIdByName(name)
		if err != nil {
			slog.Error("Error getting external application ID by name",
				slog.String("name", name),
				slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("External application not found: %v", err), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]uuid.UUID{"id": id})
	}
}

// GetExternalApplicationNameByIdHandler handles GET requests to retrieve an external application name by ID
// swagger:operation GET /external-applications/name/{ID} external-applications getExternalApplicationNameById
//
// # Get an external application name by ID
//
// ---
// produces:
// - application/json
// parameters:
//   - name: ID
//     in: path
//     required: true
//     type: string
//     format: uuid
//
// responses:
//
//	"200":
//	  description: External application name retrieved successfully
//	  schema:
//	    type: object
//	    properties:
//	      name:
//	        type: string
//	"404":
//	  description: External application not found
//	"400":
//	  description: Bad request - invalid UUID format
//	"500":
//	  description: Internal server error
func GetExternalApplicationNameByIdHandler(service ExternalApplications) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("ID")

		id, err := uuid.Parse(idStr)
		if err != nil {
			slog.Error("Error parsing UUID", slog.String("id", idStr), slog.String("error", err.Error()))
			http.Error(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		name, err := service.GetExternalApplicationNameById(id)
		if err != nil {
			slog.Error("Error getting external application name by ID",
				slog.String("id", id.String()),
				slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("External application not found: %v", err), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"name": name})
	}
}

// ExternalApplicationsHandler handles GET and POST requests for external applications
func ExternalApplicationsHandler(service ExternalApplications, authService authapi.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Require permission for reading external applications
			authapi.AuthMiddlewareRequirePermission(authService, "ReadExternalApplications", GetAllExternalApplicationsHandler(service)).ServeHTTP(w, r)
		case http.MethodPost:
			// Require permission for creating external applications
			authapi.AuthMiddlewareRequirePermission(authService, "CreateExternalApplication", CreateExternalApplicationHandler(service)).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// ExternalApplicationByIDHandler handles GET, PUT, and DELETE requests for external applications by ID
func ExternalApplicationByIDHandler(service ExternalApplications, authService authapi.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Require permission for reading external applications
			authapi.AuthMiddlewareRequirePermission(authService, "ReadExternalApplications", GetExternalApplicationByIdHandler(service)).ServeHTTP(w, r)
		case http.MethodPut:
			// Require permission for updating external applications
			authapi.AuthMiddlewareRequirePermission(authService, "UpdateExternalApplication", UpdateExternalApplicationHandler(service)).ServeHTTP(w, r)
		case http.MethodDelete:
			// Require permission for deleting external applications
			authapi.AuthMiddlewareRequirePermission(authService, "DeleteExternalApplication", DeleteExternalApplicationByIdHandler(service)).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// ExternalApplicationByNameHandler handles GET and DELETE requests for external applications by name
func ExternalApplicationByNameHandler(service ExternalApplications, authService authapi.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Require permission for reading external applications
			authapi.AuthMiddlewareRequirePermission(authService, "ReadExternalApplications", GetExternalApplicationByNameHandler(service)).ServeHTTP(w, r)
		case http.MethodDelete:
			// Require permission for deleting external applications
			authapi.AuthMiddlewareRequirePermission(authService, "DeleteExternalApplication", DeleteExternalApplicationByNameHandler(service)).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
