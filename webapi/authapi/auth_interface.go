package authapi

import (
	"github.com/babbage88/go-infra/services"
	"github.com/google/uuid"
)

type AuthService interface {
	VerifyUser(userid uuid.UUID) bool
	Login(loginReq *UserLoginRequest) UserLoginResponse
	VerifyUserPermission(executionUserId uuid.UUID, permissionsName string) (bool, error)
	CreateAuthTokenOnLogin(userid uuid.UUID, roleIds uuid.UUIDs, email string) (AuthToken, error)
	VerifyToken(tokenString string) error
	VerifyUserRolesForPermission(roleIds uuid.UUIDs, permissionName string) (bool, error)
	VerifyUserPermissionByRole(roleId uuid.UUID, permissionName string) (bool, error)
	RefreshAccessToken(refreshToken string) (AuthToken, error)
	GetUserById(id uuid.UUID) services.UserDao
}
