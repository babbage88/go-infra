package userapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/api/authapi"
	"github.com/babbage88/go-infra/services/user_crud_svc"
)

// swagger:route DELETE /role/delete RolesCRUD SoftDeleteRoleById
// Soft Delete Role by id.
//
// security:
// - bearer:
// responses:
//
// 200: SoftDeleteRoleByIdResponse
// 401: description:Unauthorized
// 403: description:Forbidden
// 404: description:Not Found
// 500: description:Internal Server Error
func SoftDeleteRoleHandleFunc(uc_service *user_crud_svc.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Extract the executing user from context
		execUserId, err := authapi.GetUserIDFromContext(r.Context())
		if err != nil {
			slog.Error("Failed to extract user ID from context", slog.String("Error", err.Error()))
			http.Error(w, fmt.Sprintf(`{"error": "Unauthorized: %s"}`, err.Error()), http.StatusUnauthorized)
			return
		}

		var request SoftDeleteRoleByIdRequest

		err = json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Verify user has DeleteRole permission
		isAdmin, err := uc_service.VerifyDeleteRole(execUserId)
		if err != nil {
			slog.Error("Error verifying user permissions", slog.String("Error", err.Error()))
			http.Error(w, fmt.Sprintf(`{"error": "Unauthorized: %s"}`, err.Error()), http.StatusUnauthorized)
			return
		}

		if !isAdmin {
			permErr := fmt.Sprintf(`{"error": "execution userId %s does not have the DeleteRole permission"}`, fmt.Sprint(execUserId))
			slog.Error("User lacks DeleteRole permission", slog.String("userId", fmt.Sprint(execUserId)))
			http.Error(w, permErr, http.StatusForbidden)
			return
		}

		response := SoftDeleteRoleByIdResponseWrapper{
			Body: SoftDeleteRoleByIdResponse{
				Error: nil,
			},
		}

		err = uc_service.SoftDeleteRoleById(request.TargetRoleId)
		if err != nil {
			slog.Error("error deleting role", slog.String("targetRole", fmt.Sprint(request.TargetRoleId)), slog.String("Error", err.Error()))
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
			return
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

func SoftDeleteRoleHandler(uc_service *user_crud_svc.UserCRUDService) http.Handler {
	return http.HandlerFunc(SoftDeleteRoleHandleFunc(uc_service))
}
