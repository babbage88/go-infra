package userapi

import (
	"github.com/babbage88/go-infra/services"
)

// Create User Request takes  in Username, Password, Email, and Role fo the new user.
// swagger:parameters idOfcreateUserEndpoint
type CreateNewUserReqWrapper struct {
	// in:body
	Body CreateNewUserRequest `json:"body"`
}

type CreateNewUserRequest struct {
	NewUsername     string `json:"newUsername"`
	NewUserPassword string `json:"newPassword"`
	NewUserEmail    string `json:"newEmail"`
	NewUserRole     string `json:"newUserRole"`
}

type GetAllUsersResponse struct {
	Users []services.UserDao `json:"users"`
}

type GetAllRolesResponse struct {
	UserRoles []services.UserRoleDao `json:"userRoles"`
}
type GetAllAppPermissionsResponse struct {
	AppPermissions []services.AppPermissionDao `json:"appPermissions"`
}

// Allows and Admin user to update another user's password.
// swagger:parameters idOfUpdateUserPw
type UpdatePasswordRequestWrapper struct {
	// in:body
	Body UpdateUserPasswordRequest `json:"body"`
}

type UpdateUserPasswordRequest struct {
	ExecutionUserId int32  `json:"executionUserId"`
	TargetUserId    int32  `json:"targetUserId"`
	NewPassword     string `json:"newPassword"`
}

type UserPasswordUpdateResponse struct {
	Success      bool  `json:"success"`
	Error        error `json:"error"`
	TargetUserId int32 `json:"targetUserId"`
}

// User Id of the executing and target users
// swagger:parameters idOfEnableUser
type EnableUserRequestWrapper struct {
	//in:body
	Body EnableUserRequest `json:"body"`
}

// User Id of the executing and target users
// swagger:parameters idOfDisableUser
type DisableUserRequestWrapper struct {
	//in:body
	Body DisableUserRequest `json:"body"`
}

type DisableUserRequest struct {
	ExecutionUserId int32 `json:"executionUserId"`
	TargetUserId    int32 `json:"targetUserId"`
}

type EnableUserRequest struct {
	ExecutionUserId int32 `json:"executionUserId"`
	TargetUserId    int32 `json:"targetUserId"`
}

type EnableDisableUserResponse struct {
	ModifiedUserInfo *services.UserDao `json:"modifiedUserInfo"`
	Error            error             `json:"error"`
}

// User Id of the executing and target users
// swagger:parameters idOfUpdateUserRole
type UpdateUserRoleMappingRequestWrapper struct {
	//in:body
	Body UpdateUserRoleMappingRequest `json:"body"`
}

type UpdateUserRoleMappingRequest struct {
	ExecutionUserId int32 `json:"executionUserId"`
	TargetUserId    int32 `json:"targetUserId"`
	RoleId          int32 `json:"roleId"`
}

type UpdateUserRoleMappingResponse struct {
	Success bool  `json:"success"`
	Error   error `json:"error"`
}

// User Id of the executing and target users
// swagger:parameters idOfCreateUserRole
type CreateUserRoleRequestWrapper struct {
	//in:body
	Body CreateUserRoleRequest `json:"body"`
}

type CreateUserRoleRequest struct {
	RoleName        string `json:"roleName"`
	RoleDescription string `json:"roleDesc"`
}

type CreateUserRoleResponse struct {
	Error           error                 `json:"error"`
	NewUserRoleInfo *services.UserRoleDao `json:"newUserRoleInfo"`
}

// Name and Description for new App Permission
// swagger:parameters idOfCreateAppPermission
type CreateAppPermissionRequestWrapper struct {
	//in: body
	Body CreateAppPermissionRequest `json:"body"`
}
type CreateAppPermissionRequest struct {
	PermissionName        string `json:"name"`
	PermissionDescription string `json:"descripiton"`
}

type CreateAppPermissionResponse struct {
	NewAppPermissionInfo *services.AppPermissionDao `json:"newPermissionInfo"`
	Error                error                      `json:"error"`
}

// Name and Description for new App Permission
// swagger:parameters idOfCreateRolePermissionMapping
type CreateRolePermissionMappingRequestWrapper struct {
	//in: body
	Body CreateRolePermissionMappingRequest `json:"body"`
}

type CreateRolePermissionMappingRequest struct {
	RoleId       int32 `json:"roleId"`
	PermissionId int32 `json:"permId"`
}

type CreateRolePermissionMappingResponse struct {
	NewMappingInfo *services.RolePermissionMappingDao `json:"newMappingInfo"`
	Error          error                              `json:"error"`
}

// Mark user as deleted in Database. Will no longer show in UI unless explicityly restored
// swagger:parameters idOfSoftDeleteUserById
type SoftDeleteUserByIdRequestWrapper struct {
	//in: body
	Body SoftDeleteUserByIdRequest `json:"body"`
}

type SoftDeleteUserByIdRequest struct {
	TargetUserId int32 `json:"targetUserId"`
}

type SoftDeleteUserByIdResponse struct {
	DeletedUserInfo *services.UserDao `json:"deletedUserInfo"`
	Error           error             `json:"error"`
}
