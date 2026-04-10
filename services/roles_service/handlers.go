package rolesservice

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// swagger:route GET /roles/permission-mappings PermissionsCRUD  GetAllAppPermissionMappings
// Returns all App Permission Mappings (which roles have which permissions).
//
// security:
// - bearer:
// responses:
//  200: GetRolePermissionMappingsResponse
// 401: description:Unauthorized
// 403: description:Forbidden
// 404: description:Not Found
// 500: description:Internal Server Error

func GetAllAppPermissionMappingsHandlerHandlerFunc(roleSvc *RoleCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var (
			response []GetRolePermissionMappingsResponse
			err      error
		)

		roleIDParam := r.URL.Query().Get("roleId")
		roleNameParam := r.URL.Query().Get("roleName")

		switch {
		case roleIDParam != "":
			roleID, parseErr := uuid.Parse(roleIDParam)
			if parseErr != nil {
				http.Error(w, "invalid roleId: "+parseErr.Error(), http.StatusBadRequest)
				return
			}
			response, err = roleSvc.GetRolePermissionMappingsByRoleID(roleID)
			if err == nil && len(response) == 0 && roleNameParam != "" {
				response, err = roleSvc.GetRolePermissionMappingsByRoleName(roleNameParam)
			}
		case roleNameParam != "":
			response, err = roleSvc.GetRolePermissionMappingsByRoleName(roleNameParam)
		default:
			response, err = roleSvc.GetRolePermissionMappings()
		}

		if err != nil {
			slog.Error("Error retrieving role permission mappings from database", slog.String("Error", err.Error()))
			http.Error(w, "Error retrieving role permission mappings from database "+err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			slog.Error("Error marshaling role permission mappings into json", slog.String("Error", err.Error()))
			http.Error(w, "Error marshaling role permission mappings to json "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

func GetAllAppPermissionMappingsHandler(roleSvc *RoleCRUDService) http.Handler {
	return http.HandlerFunc(GetAllAppPermissionMappingsHandlerHandlerFunc(roleSvc))
}
