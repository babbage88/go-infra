// Package main go-infra API.
//
// Terms Of Service:
//
// there are no TOS at this moment, use at your own risk we take no responsibility
//
//		Version: v1.1.0
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
	"log/slog"
	"os"

	"github.com/babbage88/go-infra/database/bootstrap"
	"github.com/babbage88/go-infra/internal/bumper"
	"github.com/babbage88/go-infra/internal/pretty"
	"github.com/babbage88/go-infra/services"
	"github.com/babbage88/go-infra/webapi/api_server"
	"github.com/babbage88/go-infra/webapi/authapi"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

//go:embed swagger.yaml
var swaggerSpec []byte
var versionInfo VersionInfo

//go:embed version.yaml
var versionInfoBytes []byte

func main() {
	configureDefaultLogger(slog.LevelInfo)
	versionInfo = marshalVersionInfo(versionInfoBytes)
	versionInfo.LogVersionInfo()
	var isLocalDevelopment bool
	var envFile string
	var bumpVersion bool
	var minor bool
	var major bool

	flag.BoolVar(&isLocalDevelopment, "local-development", false, "Flag to configure running local developement mode, envars set froma .env file")
	flag.StringVar(&envFile, "env-file", ".env", "Path to .env file to load Environment Variables.")
	srvport := flag.String("srvadr", ":8993", "Address and port that http server will listed on. :8993 is default")
	bootstrapNewDb := flag.Bool("db-bootstrap", false, "Create new dev database.")
	initDevUser := flag.Bool("devuser", false, "Update the devuser password")
	version := flag.Bool("version", false, "Show the current version.")
	flag.BoolVar(&bumpVersion, "bump-version", false, "Bumps version tag, push to remote repo and update version.yaml")
	flag.BoolVar(&minor, "minor", false, "Bumps Minor version number")
	flag.BoolVar(&major, "major", false, "Bumps Major version number")
	flag.Parse()

	if bumpVersion {
		var bumpErr error
		switch {
		case minor:
			bumpErr = versionInfo.FetchTagsAndBumpVersion(bumper.Minor)
		case major:
			bumpErr = versionInfo.FetchTagsAndBumpVersion(bumper.Major)
		default:
			bumpErr = versionInfo.FetchTagsAndBumpVersion(bumper.Patch)
		}
		if bumpErr != nil {
			slog.Error("Error bumping version", slog.String("error", bumpErr.Error()))
			os.Exit(1)
		}
		slog.Info("Bumped verion number", "NewVersion", versionInfo.Version)
		os.Exit(0)

	}

	if isLocalDevelopment {
		slog.Info("Local Development mode configure, loading envars from env-file", slog.String("env-file", envFile))
		err := godotenv.Load(envFile)
		if err != nil {
			slog.Error("error loading .env file", slog.String("error", err.Error()))
		}
	}

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
		versionInfo.PrintVersion()
		return
	}

	connPool := initPgConnPool()
	userService := &services.UserCRUDService{DbConn: connPool}
	authService := &authapi.LocalAuthService{DbConn: connPool}
	healthCheckService := &services.HealthCheckService{DbConn: connPool}
	if *initDevUser {
		userService.UpdateUserPasswordById(uuid.Must(uuid.Parse(os.Getenv("DEV_USER_UUID"))), os.Getenv("DEV_APP_PASS"))
	}

	api_server.StartWebApiServer(healthCheckService, authService, userService, swaggerSpec, srvport)
}
