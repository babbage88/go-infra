package user_crud_svc

import (
	"github.com/babbage88/go-infra/database/infra_db_pg"
)

type DbParser interface {
	ParseUserFromDb(dbuser infra_db_pg.User)
	ParseUserWithRoleFromDb(dbuser infra_db_pg.UsersWithRole)
	ParseUserRowFromDb(dbRow infra_db_pg.GetUserLoginRow)
	ParseAuthTokenFromDb(token infra_db_pg.ExternalAuthToken)
	ParseUserRoleFromDb(dbRow infra_db_pg.UserRole)
	ParseAppPermissionFromDb(dbRow infra_db_pg.AppPermission)
	ParseRolePermissionMappingFromDb(dbRow infra_db_pg.RolePermissionMapping)
}

func (u *UserDao) ParseUserRowFromDb(dbRow infra_db_pg.GetUserLoginRow) {
	u.Id = dbRow.ID
	u.UserName = dbRow.Username.String
	u.Email = dbRow.Email.String
	u.Enabled = dbRow.Enabled
	u.RoleIds = dbRow.RoleIds
	u.Roles = dbRow.Roles
}

func (u *UserDao) ParseUserFromDb(dbuser infra_db_pg.User) {
	u.Id = dbuser.ID
	u.UserName = dbuser.Username.String
	u.Email = dbuser.Email.String
	u.CreatedAt = dbuser.CreatedAt.Time
	u.LastModified = dbuser.LastModified.Time
	u.Enabled = dbuser.Enabled
	u.IsDeleted = dbuser.IsDeleted
}

func (u *UserDao) ParseUserWithRoleFromDb(dbuser infra_db_pg.UsersWithRole) {
	u.Id = dbuser.ID
	u.UserName = dbuser.Username.String
	u.Email = dbuser.Email.String
	u.CreatedAt = dbuser.CreatedAt.Time
	u.LastModified = dbuser.LastModified.Time
	u.Enabled = dbuser.Enabled
	u.IsDeleted = dbuser.IsDeleted
	u.RoleIds = dbuser.RoleIds
	u.Roles = dbuser.Roles
}

func (t *ExternalApplicationAuthToken) ParseAuthTokenFromDb(token infra_db_pg.ExternalAuthToken) {
	t.Id = token.ID
	t.Token = token.Token
	t.ExternalApplicationId = token.ExternalAppID
	t.UserID = token.UserID
	t.CreatedAt = token.CreatedAt.Time
	t.Expiration = token.Expiration.Time
	t.LastModified = token.LastModified.Time
}

func (t *ExternalApplication) ParseExternalApplicationFromDb(extApp infra_db_pg.ExternalIntegrationApp) {
	t.Id = extApp.ID
	t.Name = extApp.Name
}

func (ur *UserRoleDao) ParseUserRoleFromDb(dbRow infra_db_pg.UserRole) {
	ur.Id = dbRow.ID
	ur.RoleName = dbRow.RoleName
	ur.RoleDescription = dbRow.RoleDescription.String
	ur.Enabled = dbRow.Enabled
	ur.IsDeleted = dbRow.IsDeleted
	ur.CreatedAt = dbRow.CreatedAt.Time
	ur.LastModified = dbRow.LastModified.Time
}

func (ap *AppPermissionDao) ParseAppPermissionFromDb(dbRow infra_db_pg.AppPermission) {
	ap.Id = dbRow.ID
	ap.PermissionName = dbRow.PermissionName
	ap.PermissionDescription = dbRow.PermissionDescription.String
}

func (rpm *RolePermissionMappingDao) ParseRolePermissionMappingFromDb(dbRow infra_db_pg.RolePermissionMapping) {
	rpm.Id = dbRow.ID
	rpm.PermissionId = dbRow.PermissionID
	rpm.RoleId = dbRow.RoleID
	rpm.Enabled = dbRow.Enabled
	rpm.CreatedAt = dbRow.CreatedAt.Time
	rpm.LastModified = dbRow.LastModified.Time
}
