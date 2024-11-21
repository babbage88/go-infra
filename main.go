package main

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/babbage88/go-infra/services"
	"github.com/babbage88/go-infra/webapi/api_server"
	"github.com/babbage88/go-infra/webapi/authapi"
	_ "github.com/pdrum/swagger-automation/docs"
)

func main() {
	srvport := flag.String("srvadr", ":8993", "Address and port that http server will listed on. :8993 is default")
	envFilePath := flag.String("envfile", ".env", "Path to .env file to load Environment Variables.")
	username := flag.String("username", "jtrahan", "Username to create")
	pw := flag.String("pw", "", "")
	version := flag.Bool("version", false, "Show the current version.")
	testfuncs := flag.Bool("test", false, "run test module")
	flag.Parse()

	if *version {
		showVersion()
		return
	}

	envars := initEnvironment(*envFilePath)
	connPool := initPgConnPool()
	userService := &services.UserCRUDService{DbConn: connPool, Envars: envars}
	authService := &authapi.UserAuthService{DbConn: connPool, Envars: envars}

	if *testfuncs {
		request := &authapi.UserLoginRequest{UserName: *username, Password: *pw, IsHashed: false}
		loginTestResponse := request.Login(authService.DbConn)
		if loginTestResponse.Result.Success {
			slog.Info("Login Successful")
		}
		userService.NewUser(*username, *pw, fmt.Sprint(*username, "@trahan.dev"), "Admin")
		userService.UpdateUserPasswordById(1, *pw)
		return
	}

	api_server.StartWebApiServer(authService, userService, srvport)
}
