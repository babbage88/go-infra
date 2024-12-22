package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/babbage88/go-infra/auth/hashing"
	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/utils/env_helper"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserCRUDService struct {
	DbConn *pgxpool.Pool
	Envars *env_helper.EnvVars
}

type UserCRUD interface {
	NewUser(username string, hashed_pw string, email string) (UserDao, error)
	GetAllActiveUsersDao() ([]UserDao, error)
	GetAllActiveRoles([]UserRoleDao, error)
	GetAllAppPermissions([]AppPermissionDao, error)
	GetUserByName(username string) (UserDao, error)
	GetUserById(id int32) (UserDao, error)
	updateUserPasswordById(id int32, password string) error
	UpdateUserPasswordById(execUserId int32, targetUserId int32, newPassword string) error
	UpdateUserEmailById(id int32, email string)
	InsertAuthToken(token AuthTokenDao)
	VerifyAlterUser(executionUserId int32) (bool, error)
	UpdateUserPasswordWithAuth(execUserId int32, targetUserId int32, newPassword string) error
	EnableUserById(execUserId int32, targetUserId int32) (UserDao, error)
	DisableUserById(execUserId int32, targetUserId int32) (UserDao, error)
	SoftDeleteUserById(targetUserId int32) (UserDao, error)
	UpdateUserRoleMapping(execUserId int32, targetUserId int32, roleId int32) error
	CreateOrUpdateUserRole(roleName string, roleDescr string) (*UserRoleDao, error)
	CreateOrUpdateAppPermission(name string, desc string) (*AppPermissionDao, error)
	CreateOrUpdateRolePermisssionMapping(roleId int32, permId int32) (*RolePermissionMappingDao, error)
	EnableRoleById(id int32) error
	DisableRoleById(id int32) error
	SoftDeleteRoleById(id int32) error
}

func (us *UserCRUDService) UpdateUserPasswordById(execUserId int32, targetUserId int32, newPassword string) error {
	slog.Info("attempting updating user password", slog.String("execUser", fmt.Sprint(execUserId)), slog.String("targetUser", fmt.Sprint(targetUserId)))
	err := us.updateUserPasswordById(targetUserId, newPassword)
	if err != nil {
		slog.Error("error when attempting to update password", slog.String("error", err.Error()))
	}
	return err
}

func (us *UserCRUDService) UpdateUserPasswordWithAuth(execUserId int32, targetUserId int32, newPassword string) error {
	isAdmin, err := us.VerifyAlterUser(execUserId)
	if err != nil {
		slog.Error("Error Verifying user permissions.", slog.String("ID", fmt.Sprint(execUserId)), slog.String("Error", err.Error()))
		return err
	}

	if !isAdmin {
		permErr := fmt.Errorf("execution userId %d does not have the AlterUser permission", execUserId)
		return permErr
	}
	retVal := us.updateUserPasswordById(targetUserId, newPassword)
	return retVal
}

func (us *UserCRUDService) VerifyAlterUser(ueid int32) (bool, error) {
	params := infra_db_pg.VerifyUserPermissionByIdParams{
		UserId:     pgtype.Int4{Int32: ueid, Valid: true},
		Permission: pgtype.Text{String: "AlterUser", Valid: true},
	}
	queries := infra_db_pg.New(us.DbConn)
	qry, err := queries.VerifyUserPermissionById(context.Background(), params)
	if err != nil {
		slog.Error("Error verifying user permissions", slog.String("Error", err.Error()))
		return false, err
	}
	return qry, err

}

func (us *UserCRUDService) NewUser(username string, password string, email string) (UserDao, error) {
	hashed_pw, _ := hashing.HashPassword(password)
	var newuser UserDao
	// Set up parameters for the new user
	params := infra_db_pg.CreateUserParams{
		Username: pgtype.Text{String: username, Valid: true},
		Password: pgtype.Text{String: hashed_pw, Valid: true},
		Email:    pgtype.Text{String: email, Valid: true},
	}

	queries := infra_db_pg.New(us.DbConn)
	qry, err := queries.CreateUser(context.Background(), params)
	newuser.ParseUserFromDb(qry)
	return newuser, err
}

func (us *UserCRUDService) GetUserByName(username string) (*UserDao, error) {
	user := &UserDao{UserName: username}
	queries := infra_db_pg.New(us.DbConn)
	dbuser, err := queries.GetUserByName(context.Background(), pgtype.Text{username, true})
	if err != nil {
		slog.Error("error running query for username %s", username, err)
		return user, err
	}
	user.ParseUserWithRoleFromDb(dbuser)
	return user, nil
}

func (us *UserCRUDService) updateUserPasswordById(id int32, password string) error {
	hashed_pw, _ := hashing.HashPassword(password)
	params := &infra_db_pg.UpdateUserPasswordByIdParams{ID: id, Password: pgtype.Text{String: hashed_pw, Valid: true}}
	queries := infra_db_pg.New(us.DbConn)
	err := queries.UpdateUserPasswordById(context.Background(), *params)
	if err != nil {
		slog.Error("Error updating user password in database", slog.String("ID", fmt.Sprint(id)), slog.String("Error", err.Error()))
	}
	return err
}

func (us *UserCRUDService) GetUserById(id int32) (*UserDao, error) {
	user := &UserDao{Id: id}

	queries := infra_db_pg.New(us.DbConn)
	dbuser, err := queries.GetUserById(context.Background(), id)
	if err != nil {
		slog.Error("Error running query for username %d", slog.String("ID", fmt.Sprintf("%d", id)), slog.String("Error", fmt.Sprintf("%s", err)))
		return user, err
	}
	user.ParseUserWithRoleFromDb(dbuser)
	return user, nil
}

func (us *UserCRUDService) UpdateUserEmailById(id int32, email string) (*UserDao, error) {
	user := &UserDao{Id: id, Email: email}
	params := infra_db_pg.UpdateUserEmailByIdParams{ID: id, Email: pgtype.Text{String: email, Valid: true}}

	queries := infra_db_pg.New(us.DbConn)
	dbuser, err := queries.UpdateUserEmailById(context.Background(), params)
	if err != nil {
		slog.Error("Error updating email for user.", slog.String("ID", fmt.Sprintf("%d", id)), slog.String("Error", err.Error()), slog.String("Email", email))
		return user, err
	}
	user.ParseUserFromDb(dbuser)
	return user, nil
}

func (us *UserCRUDService) InsertAuthToken(t *AuthTokenDao) error {
	params := infra_db_pg.InsertAuthTokenParams{
		UserID:     pgtype.Int4{Int32: t.UserID, Valid: true},
		Token:      pgtype.Text{String: t.Token, Valid: true},
		Expiration: pgtype.Timestamp{Time: t.Expiration, InfinityModifier: 1, Valid: true},
	}

	queries := infra_db_pg.New(us.DbConn)
	err := queries.InsertAuthToken(context.Background(), params)
	if err != nil {
		slog.Error("Error inserting token.", slog.String("UserID", fmt.Sprintf("%d", t.UserID)), slog.String("Error", err.Error()))
		return err
	}
	return err
}

func (us *UserCRUDService) GetAllActiveUsersDao() ([]UserDao, error) {
	// Fetch the rows from the database
	queries := infra_db_pg.New(us.DbConn)
	rows, err := queries.GetAllActiveUsers(context.Background())
	if err != nil {
		return nil, err
	}

	// Map rows to UserDao
	userDaos := make([]UserDao, len(rows))
	for i, row := range rows {
		userDaos[i] = UserDao{
			Id:           row.ID,
			UserName:     row.Username.String,
			Email:        row.Email.String,
			Role:         row.Role,
			CreatedAt:    row.CreatedAt.Time,
			LastModified: row.LastModified.Time,
			Enabled:      row.Enabled,
			IsDeleted:    false,
		}
	}

	return userDaos, nil
}

func (us *UserCRUDService) GetAllActiveRoles() ([]UserRoleDao, error) {
	// Fetch the rows from the database
	queries := infra_db_pg.New(us.DbConn)
	rows, err := queries.GetAllUserRoles(context.Background())
	if err != nil {
		return nil, err
	}

	// Map rows to UserDao
	userRoleDaos := make([]UserRoleDao, len(rows))
	for i, row := range rows {
		userRoleDaos[i] = UserRoleDao{
			Id:              row.RoleId,
			RoleName:        row.RoleName,
			RoleDescription: row.RoleDescription.String,
			CreatedAt:       row.CreatedAt.Time,
			LastModified:    row.LastModified.Time,
			Enabled:         row.Enabled,
			IsDeleted:       false,
		}
	}

	return userRoleDaos, nil
}

func (us *UserCRUDService) GetAllAppPermissions() ([]AppPermissionDao, error) {
	// Fetch the rows from the database
	queries := infra_db_pg.New(us.DbConn)
	rows, err := queries.GetAllAppPermissions(context.Background())
	if err != nil {
		return nil, err
	}

	// Map rows to UserDao
	appPermissionDaos := make([]AppPermissionDao, len(rows))
	for i, row := range rows {
		appPermissionDaos[i] = AppPermissionDao{
			Id:                    row.ID,
			PermissionName:        row.PermissionName,
			PermissionDescription: row.PermissionDescription.String,
		}
	}

	return appPermissionDaos, nil
}

func (us *UserCRUDService) EnableUserById(execUserId int32, targetUserId int32) (*UserDao, error) {
	user := &UserDao{Id: targetUserId}

	params := infra_db_pg.EnableUserByIdParams{ID: targetUserId, Enabled: true}
	queries := infra_db_pg.New(us.DbConn)
	rows, err := queries.EnableUserById(context.Background(), params)
	if err != nil {
		user.ParseUserFromDb(rows)
		slog.Error("error enabling user", slog.String("execUser", fmt.Sprint(execUserId)), slog.String("targetUser", fmt.Sprint(targetUserId)))
		return user, err
	}
	user.ParseUserFromDb(rows)
	return user, err
}

func (us *UserCRUDService) DisableUserById(execUserId int32, targetUserId int32) (*UserDao, error) {
	user := &UserDao{Id: targetUserId}

	params := infra_db_pg.DisableUserByIdParams{ID: targetUserId, Enabled: false}
	queries := infra_db_pg.New(us.DbConn)
	rows, err := queries.DisableUserById(context.Background(), params)
	if err != nil {
		user.ParseUserFromDb(rows)
		slog.Error("error enabling user", slog.String("execUser", fmt.Sprint(execUserId)), slog.String("targetUser", fmt.Sprint(targetUserId)))
		return user, err
	}
	user.ParseUserFromDb(rows)
	return user, err
}

func (us *UserCRUDService) UpdateUserRoleMapping(execUserId int32, targetUserId int32, roleId int32) error {
	params := infra_db_pg.InsertOrUpdateUserRoleMappingByIdParams{UserID: targetUserId, RoleID: roleId}
	queries := infra_db_pg.New(us.DbConn)
	_, err := queries.InsertOrUpdateUserRoleMappingById(context.Background(), params)
	if err != nil {
		slog.Error("error modifying user group mappings", slog.String("execUser", fmt.Sprint(execUserId)), slog.String("targetUser", fmt.Sprint(targetUserId)))
		return err
	}
	return err
}

func (us *UserCRUDService) CreateOrUpdateUserRole(roleName string, roleDescr string) (*UserRoleDao, error) {
	retVal := &UserRoleDao{RoleName: roleName, RoleDescription: roleDescr}
	params := infra_db_pg.InsertOrUpdateUserRoleParams{RoleName: roleName, RoleDescription: pgtype.Text{String: roleDescr, Valid: true}}
	queries := infra_db_pg.New(us.DbConn)
	slog.Info("Executing InsertOrUpdateUserRole query", slog.String("roleName", roleName), slog.String("roleDesc", roleDescr))
	row, err := queries.InsertOrUpdateUserRole(context.Background(), params)
	if err != nil {
		slog.Error("error creating or updateing role", slog.String("roleName", roleName), slog.String("error", err.Error()))
		return retVal, err
	}
	retVal.ParseUserRoleFromDb(row)

	return retVal, err
}

func (us *UserCRUDService) EnableRoleById(id int32) error {
	queries := infra_db_pg.New(us.DbConn)
	slog.Info("Executing EnableUserRoleById Query", slog.String("id", fmt.Sprint(id)))
	err := queries.EnableUserRoleById(context.Background(), id)
	if err != nil {
		slog.Error("Error executing EnableUserRoleById Query", slog.String("error", err.Error()))
		return err
	}
	return err
}

func (us *UserCRUDService) DisableRoleById(id int32) error {
	queries := infra_db_pg.New(us.DbConn)
	slog.Info("Executing DisableUserRoleById Query", slog.String("id", fmt.Sprint(id)))
	err := queries.DisableUserRoleById(context.Background(), id)
	if err != nil {
		slog.Error("Error executing DisableUserRoleById Query", slog.String("error", err.Error()))
		return err
	}
	return err
}

func (us *UserCRUDService) SoftDeleteRoleById(id int32) error {
	queries := infra_db_pg.New(us.DbConn)
	slog.Info("Executing SoftDeleteUserRoleById Query", slog.String("id", fmt.Sprint(id)))
	err := queries.SoftDeleteUserRoleById(context.Background(), id)
	if err != nil {
		slog.Error("Error executing SoftDeleteUserRoleById Query", slog.String("error", err.Error()))
		return err
	}
	return err
}

func (us *UserCRUDService) CreateOrUpdateAppPermission(name string, desc string) (*AppPermissionDao, error) {
	retVal := &AppPermissionDao{PermissionName: name, PermissionDescription: desc}
	params := infra_db_pg.InsertOrUpdateAppPermissionParams{PermissionName: name, PermissionDescription: pgtype.Text{String: desc, Valid: true}}
	queries := infra_db_pg.New(us.DbConn)

	slog.Info("Creating App Permission", slog.String("Name", name))
	row, err := queries.InsertOrUpdateAppPermission(context.Background(), params)
	if err != nil {
		slog.Error("Error executing InsertOrUpdateAppPermission query", slog.String("error", err.Error()))
		return retVal, err
	}
	retVal.ParseAppPermissionFromDb(row)

	return retVal, err

}

func (us *UserCRUDService) CreateOrUpdateRolePermisssionMapping(roleId int32, permId int32) (*RolePermissionMappingDao, error) {
	retVal := &RolePermissionMappingDao{RoleId: roleId, PermissionId: permId}
	params := infra_db_pg.InsertOrUpdateRolePermissionMappingParams{RoleID: roleId, PermissionID: permId}
	queries := infra_db_pg.New(us.DbConn)

	slog.Info("Creating Role Permission Mapping", slog.String("RoleId", fmt.Sprint(roleId)), slog.String("PermissionId", fmt.Sprint(permId)))
	row, err := queries.InsertOrUpdateRolePermissionMapping(context.Background(), params)
	if err != nil {
		slog.Error("Error executing InsertOrUpdateRolePermissionMapping query", slog.String("error", err.Error()))
		return retVal, err
	}
	retVal.ParseRolePermissionMappingFromDb(row)

	return retVal, err
}

func (us *UserCRUDService) SoftDeleteUserById(targetUserId int32) (*UserDao, error) {
	retVal := &UserDao{Id: targetUserId}
	queries := infra_db_pg.New(us.DbConn)

	slog.Info("Executing SofDeleteUserById", slog.String("targetUserId", fmt.Sprint(targetUserId)))
	row, err := queries.SoftDeleteUserById(context.Background(), targetUserId)
	if err != nil {
		slog.Error("Error deleteing user", slog.String("error", err.Error()))
		return retVal, err
	}
	retVal.ParseUserFromDb(row)

	return retVal, err
}
