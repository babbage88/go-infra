package userapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/services"
)

// swagger:route POST /create/user createuser idOfcreateUserEndpoint
// Create a new user.
//
// security:
// - bearer:
// responses:
//   200: UserDao

func CreateUser(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			slog.Error("Invalid request method", slog.String("Method", r.Method))
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		var newUserReq CreateNewUserRequest

		err := json.NewDecoder(r.Body).Decode(&newUserReq)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		newUser, err := uc_service.NewUser(
			newUserReq.NewUsername,
			newUserReq.NewUserPassword,
			newUserReq.NewUserEmail)
		if err != nil {
			slog.Error("Error creating new user", slog.String("Error", err.Error()))
			http.Error(w, "Error createing new user "+err.Error(), http.StatusInternalServerError)
		}

		jsonResponse, err := json.Marshal(newUser)
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

// swagger:route POST /update/userpass updateUserPw idOfUpdateUserPw
// Update user password.
//
// security:
// - bearer:
// responses:
//
//	200: UserPasswordUpdateResponse
func UpdateUserPassword(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request UpdateUserPasswordRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		response := UserPasswordUpdateResponse{
			TargetUserId: request.TargetUserId,
		}
		response.Error = uc_service.UpdateUserPasswordById(request.ExecutionUserId, request.TargetUserId, request.NewPassword)
		if response.Error != nil {
			response.Success = false
			http.Error(w, "error updating user password "+err.Error(), http.StatusUnauthorized)
			return
		}
		response.Success = true

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

// swagger:route POST /users getallusers idOfgetAllUsersEndpoint
// Returns all active users.
//
// security:
// - bearer:
// responses:
//   200: GetAllUsersResponse

func GetAllUsers(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		users, err := uc_service.GetAllActiveUsersDao()
		if err != nil {
			slog.Error("Error getting users from database", slog.String("Error", err.Error()))
			http.Error(w, "Error createing new user "+err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse, err := json.Marshal(users)
		if err != nil {
			slog.Error("Error marshaling users into json", slog.String("Error", err.Error()))
			http.Error(w, "Error marshaling users to json "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

// swagger:route POST /user/enable enableUser idOfEnableUser
// Enable specified target User Id.
//
// security:
// - bearer:
// responses:
//
//	200: EnableDisableUserResponse
func EnableUser(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request EnableUserRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		modifiedUserInfo := &services.UserDao{Id: request.TargetUserId}
		response := EnableDisableUserResponse{
			ModifiedUserInfo: modifiedUserInfo,
			Error:            err,
		}

		response.ModifiedUserInfo, response.Error = uc_service.EnableUserById(request.ExecutionUserId, request.TargetUserId)
		if response.Error != nil {
			http.Error(w, "error enabling user password "+response.Error.Error(), http.StatusUnauthorized)
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

// swagger:route POST /user/disable disableUser idOfDisableUser
// Disable specified target User Id.
//
// security:
// - bearer:
// responses:
//
//	200: UpdateUserRoleResponse
func DisableUser(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request DisableUserRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		modifiedUserInfo := &services.UserDao{Id: request.TargetUserId}
		response := EnableDisableUserResponse{
			ModifiedUserInfo: modifiedUserInfo,
			Error:            err,
		}

		response.ModifiedUserInfo, response.Error = uc_service.DisableUserById(request.ExecutionUserId, request.TargetUserId)
		if response.Error != nil {
			http.Error(w, "error enabling user password "+response.Error.Error(), http.StatusUnauthorized)
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

// swagger:route POST /user/role updateUserRole idOfUpdateUserRole
// Update User Role Mapping
//
// security:
// - bearer:
// responses:
//
//	200: EnableDisableUserResponse
func UpdateUserRole(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request UpdateUserRoleRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		response := UpdateUserRoleResponse{
			Error: err,
		}

		response.Error = uc_service.UpdateUserRole(request.ExecutionUserId, request.TargetUserId, request.RoleId)
		if response.Error != nil {
			http.Error(w, "error updating user role "+response.Error.Error(), http.StatusUnauthorized)
			return
		} else {
			response.Success = true
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
