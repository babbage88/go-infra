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
	"os"

	"github.com/babbage88/go-infra/database/bootstrap"
	"github.com/babbage88/go-infra/internal/pretty"
	"github.com/babbage88/go-infra/services"
	"github.com/babbage88/go-infra/webapi/api_server"
	"github.com/babbage88/go-infra/webapi/authapi"
)

//go:embed swagger.yaml
var swaggerSpec []byte

func main() {
	srvport := flag.String("srvadr", ":8993", "Address and port that http server will listed on. :8993 is default")
	envFilePath := flag.String("envfile", ".env", "Path to .env file to load Environment Variables.")
	bootstrapNewDb := flag.Bool("db-bootstrap", false, "Create new dev database.")
	initDevUser := flag.Bool("devuser", false, "Update the devuser password")
	version := flag.Bool("version", false, "Show the current version.")
	flag.Parse()

	envars := initEnvironment(*envFilePath)

	if *bootstrapNewDb {
		bootstrap.NewDb()
		pretty.Print("test")
		err := bootstrap.CreateInfradbUser(os.Getenv("DB_USER"))
		if err != nil {
			pretty.PrintErrorf("Error configuring db user %s", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *version {
		showVersion()
		return
	}

	connPool := initPgConnPool()
	userService := &services.UserCRUDService{DbConn: connPool, Envars: envars}
	authService := &authapi.UserAuthService{DbConn: connPool, Envars: envars}
	healthCheckService := &services.HealthCheckService{DbConn: connPool, Envars: envars}
	if *initDevUser {
		userService.UpdateUserPasswordById(1, envars.GetVarMapValue("DEV_APP_PASS"))
	}

	api_server.StartWebApiServer(healthCheckService, authService, userService, swaggerSpec, srvport)
}
