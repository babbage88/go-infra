package main

import (
	"fmt"

	infra_db "github.com/babbage88/go-infra/database/infra_db"
	env_helper "github.com/babbage88/go-infra/utils/env_helper"
	"github.com/jackc/pgx/v5/pgxpool"
)

func initEnvironment(envPath string) *env_helper.EnvVars {
	env_helper.LoadEnvFile(*&envPath)
	envars := env_helper.NewDotEnvSource(env_helper.WithDotEnvFileName(*&envPath))
	fmt.Printf("EnVars file name: %s\n", envars.DotFileName)
	envars.ParseEnvVariables()
	return envars
}

func initPgConnPool() *pgxpool.Pool {
	connPool := infra_db.PgPoolInit()
	return connPool
}
