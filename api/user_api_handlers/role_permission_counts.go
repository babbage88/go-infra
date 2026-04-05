package userapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/services/user_crud_svc"
)

// swagger:route GET /roles/permission-counts RolesCRUD GetRolesPermissionCounts
// Get permission counts for all roles.
//
// security:
// - bearer:
// responses:
//
// 200: GetRolesPermissionCountResponse
// 401: description:Unauthorized
// 500: description:Internal Server Error
func GetRolesPermissionCountsHandleFunc(uc_service *user_crud_svc.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		rolePermissionCounts, err := uc_service.GetRolesPermissionCounts()
		if err != nil {
			slog.Error("error fetching role permission counts", slog.String("Error", err.Error()))
			response := GetRolesPermissionCountResponseWrapper{
				Body: GetRolesPermissionCountResponse{
					RolePermissionCounts: []RolePermissionCount{},
					Error:                err,
				},
			}
			jsonResponse, _ := json.Marshal(response)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonResponse)
			return
		}

		// Convert database results to API response
		rpcList := make([]RolePermissionCount, 0, len(rolePermissionCounts))
		for _, rpc := range rolePermissionCounts {
			rpcList = append(rpcList, RolePermissionCount{
				RoleId:          rpc.ID,
				RoleName:        rpc.RoleName,
				PermissionCount: rpc.PermissionCount,
			})
		}

		response := GetRolesPermissionCountResponseWrapper{
			Body: GetRolesPermissionCountResponse{
				RolePermissionCounts: rpcList,
				Error:                nil,
			},
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			slog.Error("Failed to marshal JSON response", slog.String("Error", err.Error()))
			http.Error(w, "Failed to marshal JSON response: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

func GetRolesPermissionCountsHandler(uc_service *user_crud_svc.UserCRUDService) http.Handler {
	return http.HandlerFunc(GetRolesPermissionCountsHandleFunc(uc_service))
}
