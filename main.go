package main

import (
	"flag"
	"log/slog"

	"github.com/babbage88/go-infra/auth/hashing"
	infra_db "github.com/babbage88/go-infra/database/infra_db"
	test "github.com/babbage88/go-infra/utils/test"
	"github.com/babbage88/go-infra/webapi/api_server"
	_ "github.com/pdrum/swagger-automation/docs"
)

func main() {

	srvport := flag.String("srvadr", ":8993", "Address and port that http server will listed on. :8993 is default")
	envFilePath := flag.String("envfile", ".env", "Path to .env file to load Environment Variables.")
	username := flag.String("username", "jtrahan", "Username to create")
	emailuser := flag.String("email", "testdev@trahan.dev", "email for new username")
	version := flag.Bool("version", false, "Show the current version.")
	testfuncs := flag.Bool("test", false, "run test module")
	flag.Parse()

	if *version {
		showVersion()
		return
	}

	envars := initEnvironment(*envFilePath)
	connPool := initPgConnPool()

	if *testfuncs {
		login_pw := envars.GetVarMapValue("DEV_APP_TEST_PW")
		hashed_pw, err := hashing.HashPassword(login_pw)
		if err != nil {
			slog.Info("Error hashing password.")
		}
		test.TestCreateNewUser(connPool, *username, hashed_pw, *emailuser, "admin")
	}

	db := initializeDbConn(envars)

	defer func() {
		if err := infra_db.CloseDbConnection(); err != nil {
			slog.Error("Failed to close the database connection: ", "Error", err)
		}
	}()

	api_server.StartWebApiServer(envars, db, srvport)

}
