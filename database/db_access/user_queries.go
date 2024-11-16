package dbaccess

import (
	"context"
	"log/slog"

	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUserQuery(connPool *pgxpool.Pool, username string, hashed_pw string, email string, role string) (infra_db_pg.User, error) {
	// Set up parameters for the new user
	params := infra_db_pg.CreateUserParams{
		Username: pgtype.Text{String: username, Valid: true},
		Password: pgtype.Text{String: hashed_pw, Valid: true},
		Email:    pgtype.Text{String: email, Valid: true},
		Role:     pgtype.Text{String: role, Valid: true},
	}

	queries := infra_db_pg.New(connPool)
	newUser, err := queries.CreateUser(context.Background(), params)
	return newUser, err
}

type DbParser interface {
	//ParseUserFromDb(dbuser infra_db_pg.User) (UserDao, error)
	ParseUserFromDb(dbuser infra_db_pg.User)
	ParseAuthTokenFromDb(token infra_db_pg.AuthToken)
}

func (u *UserDao) ParseUserFromDb(dbuser infra_db_pg.User) {
	u.Id = dbuser.ID
	u.UserName = dbuser.Username.String
	u.Password = dbuser.Password.String
	u.Email = dbuser.Email.String
	u.Role = dbuser.Role.String
	u.CreatedAt = dbuser.CreatedAt.Time
	u.LastModified = dbuser.LastModified.Time
	u.Enabled = dbuser.Enabled
	u.IsDeleted = dbuser.IsDeleted
}

func (t *AuthTokenDao) ParseAuthTokenFromDb(token infra_db_pg.AuthToken) {
	t.Id = token.ID
	t.Token = token.Token.String
	t.UserID = token.UserID.Int32
	t.CreatedAt = token.CreatedAt.Time
	t.Expiration = token.Expiration.Time
	t.LastModified = token.LastModified.Time
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

type UserActions interface {
	VerifyUserPassword(connPool *pgxpool.Pool) bool
	GetUserDaoByUsername(connPool *pgxpool.Pool, username string) (UserDao, error)
	GetUserDaoById(connPool *pgxpool.Pool, id int32) (UserDao, error)
}
