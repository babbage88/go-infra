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
	NewUser(username string, hashed_pw string, email string, role string) (UserDao, error)
	GetUserByName(username string) (UserDao, error)
	GetUserById(id int32) (UserDao, error)
	updateUserPasswordById(id int32, password string) error
	UpdateUserEmailById(id int32, email string)
	InsertAuthToken(token AuthTokenDao)
	VerifyAlterUser(executionUserId int32) (bool, error)
	UpdateUserPasswordWithAuth(execUserId int32, targetUserId int32, newPassword string) error
}

func (us *UserCRUDService) UpdateUserPasswordWithAuth(execUserId int32, targetUserId int32, newPassword string) error {
	isAdmin, err := us.VerifyAlterUser(execUserId)
	if err != nil {
		slog.Error("Error Verifying user permissions.", slog.String("ID", fmt.Sprint(execUserId)), slog.String("Error", err.Error()))
		return err
	}

	if !isAdmin {
		permErr := fmt.Errorf("Execution userId %d does not have the AlterUser permission", execUserId)
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

func (us *UserCRUDService) NewUser(username string, password string, email string, role string) (UserDao, error) {
	hashed_pw, _ := hashing.HashPassword(password)
	var newuser UserDao
	// Set up parameters for the new user
	params := infra_db_pg.CreateUserParams{
		Username: pgtype.Text{String: username, Valid: true},
		Password: pgtype.Text{String: hashed_pw, Valid: true},
		Email:    pgtype.Text{String: email, Valid: true},
		Role:     pgtype.Text{String: role, Valid: true},
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
		slog.Error("Eroor running query for username %s", username, err)
		return user, err
	}
	user.ParseUserFromDb(dbuser)
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
	user.ParseUserFromDb(dbuser)
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
