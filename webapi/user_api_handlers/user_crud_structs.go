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
