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

	"github.com/babbage88/go-infra/services"
	"github.com/babbage88/go-infra/webapi/api_server"
	"github.com/babbage88/go-infra/webapi/authapi"
)

//go:embed swagger.yaml
var swaggerSpec []byte

func main() {
	srvport := flag.String("srvadr", ":8993", "Address and port that http server will listed on. :8993 is default")
	envFilePath := flag.String("envfile", ".env", "Path to .env file to load Environment Variables.")
	version := flag.Bool("version", false, "Show the current version.")
	flag.Parse()

	if *version {
		showVersion()
		return
	}

	envars := initEnvironment(*envFilePath)
	connPool := initPgConnPool()
	userService := &services.UserCRUDService{DbConn: connPool, Envars: envars}
	authService := &authapi.UserAuthService{DbConn: connPool, Envars: envars}

	api_server.StartWebApiServer(authService, userService, swaggerSpec, srvport)
}
