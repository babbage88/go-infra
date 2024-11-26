// Package main go-infra API.
//
// Terms Of Service:
//
// there are no TOS at this moment, use at your own risk we take no responsibility
//
//		Version: v1.0.7
//		License: N/A
//		Contact: Justin Trahan<test@trahan.dev>
//
//		Consumes:
//		- application/json
//
//		Produces:
//		- application/json
//
//	    Security:
//	    - bearer:
//
//	    SecurityDefinitions:
//	      bearer:
//	         type: apiKey
//	         name: Authorization
//	         in: header
//
// swagger:meta
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log/slog"

	"github.com/babbage88/go-infra/services"
	"github.com/babbage88/go-infra/webapi/api_server"
	"github.com/babbage88/go-infra/webapi/authapi"
	_ "github.com/pdrum/swagger-automation/docs"
)

//go:embed swagger.yaml
var swaggerSpec []byte

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
		request := &authapi.UserLoginRequest{UserName: *username, Password: *pw}
		loginTestResponse := request.Login(authService.DbConn)
		if loginTestResponse.Result.Success {
			slog.Info("Login Successful")
		}
		userService.NewUser(*username, *pw, fmt.Sprint(*username, "@trahan.dev"), "Admin")
		userService.UpdateUserPasswordById(1, *pw)
		return
	}

	api_server.StartWebApiServer(authService, userService, swaggerSpec, srvport)
}
