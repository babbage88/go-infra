package main

import (
	infra_db "github.com/babbage88/go-infra/database/infra_db"
	"github.com/jackc/pgx/v5/pgxpool"
)

func initPgConnPool() *pgxpool.Pool {
	connPool := infra_db.PgPoolInit()
	return connPool
}
