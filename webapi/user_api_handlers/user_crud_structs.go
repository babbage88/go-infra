package userapi

import (
	"github.com/babbage88/go-infra/services/user_crud_svc"
	"github.com/google/uuid"
)

// Create User Request takes  in Username, Password, Email, and Role fo the new user.
// swagger:parameters CreateUser
type CreateNewUserReqWrapper struct {
	// in:body
	Body CreateNewUserRequest `json:"body"`
}

type CreateNewUserRequest struct {
	NewUsername     string `json:"newUsername"`
	NewUserPassword string `json:"newPassword"`
	NewUserEmail    string `json:"newEmail"`
}

// swagger:response CreateUserResponse
type CreateUserResponseWrapper struct {
	// in: body
	Body user_crud_svc.UserDao `json:"body"`
}

// swagger:response GetAllUsersResponse
type GetAllUsersResponse struct {
	Users []user_crud_svc.UserDao `json:"users"`
}

// swagger:parameters getUserById
type GetUserByIdRequest struct {
	// ID of user
	//
	// In: path
	ID string `json:"ID"`
}

// swagger:response GetUserByIdResponse
type GetUserByIdResponseWrapper struct {
	// in: body
	Body GetUserByIdResponse `json:"body"`
}

type GetUserByIdResponse struct {
	User user_crud_svc.UserDao `json:"user"`
}

type GetAllRolesResponseWrapper struct {
	Body GetAllRolesResponse `json:"body"`
}

// swagger:response GetAllRolesResponse
type GetAllRolesResponse struct {
	UserRoles []user_crud_svc.UserRoleDao `json:"userRoles"`
}

type GetAllAppPermissionsResponse struct {
	AppPermissions []user_crud_svc.AppPermissionDao `json:"appPermissions"`
}

// Allows and Admin user to update another user's password.
// swagger:parameters UpdateUserPw
type UpdatePasswordRequestWrapper struct {
	// in: body
	Body UpdateUserPasswordRequest `json:"body"`
}

type UpdateUserPasswordRequest struct {
	TargetUserId uuid.UUID `json:"targetUserId"`
	NewPassword  string    `json:"newPassword"`
}

// swagger:response UserPasswordUpdateResponse
type UserPasswordUpdateResponseWrapper struct {
	// in: body
	Body UserPasswordUpdateResponse `json:"body"`
}

type UserPasswordUpdateResponse struct {
	Success      bool      `json:"success"`
	Error        error     `json:"error"`
	TargetUserId uuid.UUID `json:"targetUserId"`
}

// User Id of the executing and target users
// swagger:parameters EnableUser
type EnableUserRequestWrapper struct {
	//in: body
	Body EnableUserRequest `json:"body"`
}

// User Id of the executing and target users
// swagger:parameters DisableUser
type DisableUserRequestWrapper struct {
	//in:body
	Body DisableUserRequest `json:"body"`
}

type DisableUserRequest struct {
	TargetUserId uuid.UUID `json:"targetUserId"`
}

type EnableUserRequest struct {
	TargetUserId uuid.UUID `json:"targetUserId"`
}

// swagger:response EnableDisableUserResponse
type EnableDisableUserResponseWrapper struct {
	// in: body
	Body EnableDisableUserResponse `json:"body"`
}
type EnableDisableUserResponse struct {
	ModifiedUserInfo *user_crud_svc.UserDao `json:"modifiedUserInfo"`
	Error            error                  `json:"error"`
}

// User Id of the executing and target users
// swagger:parameters UpdateUserRole DisableUserRoleMapping
type UpdateUserRoleMappingRequestWrapper struct {
	// in: body
	Body UpdateUserRoleMappingRequest `json:"body"`
}

type UpdateUserRoleMappingRequest struct {
	TargetUserId uuid.UUID `json:"targetUserId"`
	RoleId       uuid.UUID `json:"roleId"`
}

// swagger:response UpdateUserRoleMappingResponse
type UpdateUserRoleMappingResponseWrapper struct {
	// in: body
	Body UpdateUserRoleMappingResponse `json:"body"`
}

// swagger:model UpdateUserRoleMappingResponse
type UpdateUserRoleMappingResponse struct {
	Success bool  `json:"success"`
	Error   error `json:"error"`
}

// User Id of the executing and target users
// swagger:parameters CreateUserRole
type CreateUserRoleRequestWrapper struct {
	// in:body
	Body CreateUserRoleRequest `json:"body"`
}

type CreateUserRoleRequest struct {
	RoleName        string `json:"roleName"`
	RoleDescription string `json:"roleDesc"`
}

// swagger:response CreateUserRoleResponse
type CreateUserRoleResponseWrapper struct {
	// in: body
	Body CreateUserRoleResponse `json:"body"`
}

// swagger:model UserRoleDao
type CreateUserRoleResponse struct {
	Error           error                      `json:"error"`
	NewUserRoleInfo *user_crud_svc.UserRoleDao `json:"newUserRoleInfo"`
}

// Name and Description for new App Permission
// swagger:parameters CreateAppPermission
type CreateAppPermissionRequestWrapper struct {
	//in: body
	Body CreateAppPermissionRequest `json:"body"`
}
type CreateAppPermissionRequest struct {
	PermissionName        string `json:"name"`
	PermissionDescription string `json:"descripiton"`
}

// swagger:response CreateAppPermissionResponse
type CreateAppPermissionResponseWrapper struct {
	// in: body
	Body CreateAppPermissionResponse `json:"body"`
}

// swagger:model AppPermissionDao
type CreateAppPermissionResponse struct {
	NewAppPermissionInfo *user_crud_svc.AppPermissionDao `json:"newPermissionInfo"`
	Error                error                           `json:"error"`
}

// Name and Description for new App Permission
// swagger:parameters CreateRolePermissionMapping
type CreateRolePermissionMappingRequestWrapper struct {
	//in: body
	Body CreateRolePermissionMappingRequest `json:"body"`
}

type CreateRolePermissionMappingRequest struct {
	RoleId       uuid.UUID `json:"roleId"`
	PermissionId uuid.UUID `json:"permId"`
}

// swagger:response CreateRolePermissionMappingResponse
type CreateRolePermissionMappingResponseWrapper struct {
	// in: body
	Body CreateRolePermissionMappingResponse `json:"body"`
}

// swagger:model RolePermissionMappingDao
type CreateRolePermissionMappingResponse struct {
	NewMappingInfo *user_crud_svc.RolePermissionMappingDao `json:"newMappingInfo"`
	Error          error                                   `json:"error"`
}

// Mark user as deleted in Database. Will no longer show in UI unless explicityly restored
// swagger:parameters SoftDeleteUserById
type SoftDeleteUserByIdRequestWrapper struct {
	//in: body
	Body SoftDeleteUserByIdRequest `json:"body"`
}

type SoftDeleteUserByIdRequest struct {
	TargetUserId uuid.UUID `json:"targetUserId"`
}

// swagger:response SoftDeleteUserByIdResponse
type SoftDeleteUserByIdResponseWrapper struct {
	// in: body
	Body SoftDeleteUserByIdResponse `json:"body"`
}

type SoftDeleteUserByIdResponse struct {
	DeletedUserInfo *user_crud_svc.UserDao `json:"deletedUserInfo"`
	Error           error                  `json:"error"`
}
