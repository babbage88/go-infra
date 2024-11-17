package db_access

import (
	"context"
	"log/slog"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	//"github.com/babbage88/go-infra/webapi/authapi"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUserQuery(connPool *pgxpool.Pool, username string, hashed_pw string, email string, role string) (UserDao, error) {
	var newuser UserDao
	// Set up parameters for the new user
	params := infra_db_pg.CreateUserParams{
		Username: pgtype.Text{String: username, Valid: true},
		Password: pgtype.Text{String: hashed_pw, Valid: true},
		Email:    pgtype.Text{String: email, Valid: true},
		Role:     pgtype.Text{String: role, Valid: true},
	}

	queries := infra_db_pg.New(connPool)
	qry, err := queries.CreateUser(context.Background(), params)
	newuser.ParseUserFromDb(qry)
	return newuser, err
}

type UserCRUDActions interface {
	GetUserDaoByUsername(connPool *pgxpool.Pool, username string) (UserDao, error)
	GetUserDaoById(connPool *pgxpool.Pool, id int32) (UserDao, error)
}

func GetUserDaoByUsername(connPool *pgxpool.Pool, username string) (*UserDao, error) {
	var user UserDao
	queries := infra_db_pg.New(connPool)
	dbuser, err := queries.GetUserByName(context.Background(), pgtype.Text{username, true})
	if err != nil {
		slog.Error("Eroor running query for username %s", username, err)
		return &user, err
	}
	user.ParseUserFromDb(dbuser)
	return &user, nil
}

func GetUserDaoById(connPool *pgxpool.Pool, id int32) (*UserDao, error) {
	var user UserDao

	queries := infra_db_pg.New(connPool)
	dbuser, err := queries.GetUserById(context.Background(), id)
	if err != nil {
		slog.Error("Eroor running query for username %s", id, err)
		return &user, err
	}
	user.ParseUserFromDb(dbuser)
	return &user, nil
}
