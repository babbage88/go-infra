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
