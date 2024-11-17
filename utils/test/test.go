package test

import (
	"log"

	"github.com/babbage88/go-infra/database/db_access"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCreateNewUser(connPool *pgxpool.Pool, username string, hashed_pw string, email string, role string) (db_access.UserDao, error) {
	newuser, err := db_access.CreateUserQuery(connPool, username, hashed_pw, email, role)
	if err != nil {
		log.Fatalf("Error creating user %s", err)
	}
	return newuser, err
}
