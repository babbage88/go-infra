package userapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/services"
	"github.com/babbage88/go-infra/utils/type_helper"
)

// swagger:route POST /create/user createuser idOfcreateUserEndpoint
// Create a new user.
//
// security:
// - bearer:
// responses:
//   200: UserDao

func CreateUserHandler(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
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
func UpdateUserPasswordHandler(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
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
		response.Error = uc_service.UpdateUserPasswordById(request.TargetUserId, request.NewPassword)
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

// swagger:route GET /users getallusers idOfgetAllUsersEndpoint
// Returns all active users.
//
// security:
// - bearer:
// responses:
//   200: GetAllUsersResponse

func GetAllUsersHandler(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		users, err := uc_service.GetAllActiveUsersDao()
		if err != nil {
			slog.Error("Error getting users from database", slog.String("Error", err.Error()))
			http.Error(w, "Error getting users from database "+err.Error(), http.StatusInternalServerError)
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

// swagger:route GET /users/{ID} getUserById idOfgetUserByIdEndpoint
// Returns User Info for the user id specified in URL users.
//
// security:
// - bearer:
// responses:
//   200: GetUserByIdResponse

func GetUserByIdHandler(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		urlId := r.PathValue("ID")
		id, err := type_helper.ParseInt32(urlId)
		if err != nil {
			slog.Error("Error Parsing user id from URL path", slog.String("Error", err.Error()))
			http.Error(w, "Error Parsing user id from URL path "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		slog.Info("Getting user info for User Id", slog.String("ID", urlId))
		user, err := uc_service.GetUserById(id)
		if err != nil {
			slog.Error("Error getting user from database", slog.String("Error", err.Error()))
			http.Error(w, "Error getting user from database "+err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse, err := json.Marshal(user)
		if err != nil {
			slog.Error("Error marshaling user into json", slog.String("Error", err.Error()))
			http.Error(w, "Error marshaling user to json "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

// swagger:route GET /roles getAllUserRoles idOfgetAllRolesEndpoint
// Returns all active User Roles.
//
// security:
// - bearer:
// responses:
//   200: GetAllRolesResponse

func GetAllRolesHandler(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		roles, err := uc_service.GetAllActiveRoles()
		if err != nil {
			slog.Error("Error getting user roles from database", slog.String("Error", err.Error()))
			http.Error(w, "Error getting user roles from database "+err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse, err := json.Marshal(roles)
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

// swagger:route GET /permissions getAllAppPermissions idOfgetAllAppPermissionsEndpoint
// Returns all App Permissions
//
// security:
// - bearer:
// responses:
//   200: GetAllAppPermissionsResponse

func GetAllAppPermissionsHandler(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		appPermissions, err := uc_service.GetAllAppPermissions()
		if err != nil {
			slog.Error("Error retrieving app permissions from database", slog.String("Error", err.Error()))
			http.Error(w, "Error retrieving app permissions from database "+err.Error(), http.StatusInternalServerError)
			return
		}
		jsonResponse, err := json.Marshal(appPermissions)
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
func EnableUserHandler(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
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

		response.ModifiedUserInfo, response.Error = uc_service.EnableUserById(request.TargetUserId)
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
func DisableUserHandler(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
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

		response.ModifiedUserInfo, response.Error = uc_service.DisableUserById(request.TargetUserId)
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
func UpdateUserRoleMappingHandler(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request UpdateUserRoleMappingRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		response := UpdateUserRoleMappingResponse{
			Error: err,
		}

		slog.Info("Updating user role mapping", slog.String("targetUserID", fmt.Sprint(request.TargetUserId)), slog.String("roleID", fmt.Sprint(request.RoleId)))
		response.Error = uc_service.UpdateUserRoleMapping(request.TargetUserId, request.RoleId)
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

// swagger:route POST /create/role createUserRole idOfCreateUserRole
// Create New User Role.
//
// security:
// - bearer:
// responses:
//
//	200: CreateUserRoleResponse
func CreateUserRoleHandler(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request CreateUserRoleRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		newUserRoleInfo := &services.UserRoleDao{RoleName: request.RoleName, RoleDescription: request.RoleDescription}
		response := CreateUserRoleResponse{
			NewUserRoleInfo: newUserRoleInfo,
			Error:           err,
		}

		slog.Info("CreateorUpdate user role", slog.String("RoleName", request.RoleName))
		response.NewUserRoleInfo, response.Error = uc_service.CreateOrUpdateUserRole(request.RoleName, request.RoleDescription)
		if response.Error != nil {
			http.Error(w, "error creating role "+response.Error.Error(), http.StatusUnauthorized)
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

// swagger:route POST /create/permission createAppPermission idOfCreateAppPermission
// Create New App Permission.
//
// security:
// - bearer:
// responses:
//
//	200: CreateAppPermissionResponse
func CreateAppPermissionHandler(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request CreateAppPermissionRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		newAppPermissionInfo := &services.AppPermissionDao{PermissionName: request.PermissionName, PermissionDescription: request.PermissionDescription}
		response := CreateAppPermissionResponse{
			NewAppPermissionInfo: newAppPermissionInfo,
			Error:                err,
		}

		response.NewAppPermissionInfo, response.Error = uc_service.CreateOrUpdateAppPermission(request.PermissionName, request.PermissionDescription)
		if response.Error != nil {
			http.Error(w, "error creating role "+response.Error.Error(), http.StatusUnauthorized)
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

// swagger:route POST /roles/permission createRolePermissionMapping idOfCreateRolePermissionMapping
// Map App Permission to User Role.
//
// security:
// - bearer:
// responses:
//
//	200: CreateRolePermissionMapptingResponse
func CreateRolePermissionMappingHandler(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request CreateRolePermissionMappingRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		newRolePermissionMappingInfo := &services.RolePermissionMappingDao{RoleId: request.RoleId, PermissionId: request.PermissionId}
		response := CreateRolePermissionMappingResponse{
			NewMappingInfo: newRolePermissionMappingInfo,
			Error:          err,
		}

		response.NewMappingInfo, response.Error = uc_service.CreateOrUpdateRolePermisssionMapping(request.RoleId, request.PermissionId)
		if response.Error != nil {
			http.Error(w, "error creating role permission mapping "+response.Error.Error(), http.StatusUnauthorized)
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

// swagger:route POST /user/delete SoftDeleteUserHandler idOfSoftDeleteUserById
// Soft Delete User by id.
//
// security:
// - bearer:
// responses:
//
//	200: SoftDeleteUserByIdResponse
func SoftDeleteUserHandler(uc_service *services.UserCRUDService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var request SoftDeleteUserByIdRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			slog.Error("Failed to decode request body", slog.String("Error", err.Error()))
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		deletedUserInfo := &services.UserDao{Id: request.TargetUserId}
		response := SoftDeleteUserByIdResponse{
			DeletedUserInfo: deletedUserInfo,
			Error:           err,
		}

		response.DeletedUserInfo, response.Error = uc_service.SoftDeleteUserById(request.TargetUserId)
		if response.Error != nil {
			http.Error(w, "error deleting user "+response.Error.Error(), http.StatusUnauthorized)
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
