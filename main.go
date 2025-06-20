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
	"log/slog"
	"os"

	"github.com/babbage88/go-infra/api/api_server"
	"github.com/babbage88/go-infra/api/authapi"
	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/services/external_applications"
	"github.com/babbage88/go-infra/services/host_servers"
	"github.com/babbage88/go-infra/services/ssh_key_provider"
	"github.com/babbage88/go-infra/services/user_crud_svc"
	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/google/uuid"
)

//go:embed swagger.yaml
var swaggerSpec []byte
var versionInfo VersionInfo

//go:generate go tool oapi-codegen -config cfg.yaml swagger.yaml

//go:embed version.yaml
var versionInfoBytes []byte

func main() {
	// initialize default logger and configure slog with tint
	configureDefaultLogger(slog.LevelInfo)

	// marshal version information from embeded version.yaml
	versionInfo = marshalVersionInfo(versionInfoBytes)
	versionInfo.LogVersionInfo()

	// parse flags and execute any special statup functions.
	parseFlags()
	configureStartupOptions()

	connPool := initPgConnPool()
	userService := &user_crud_svc.UserCRUDService{DbConn: connPool}
	authService := &authapi.LocalAuthService{DbConn: connPool}
	healthCheckService := &user_crud_svc.HealthCheckService{DbConn: connPool}
	secretProvider := &user_secrets.PgUserSecretStore{DbConn: connPool}
	hostServerProvider := host_servers.NewHostServerProvider(infra_db_pg.New(connPool))
	sshKeyProvider := ssh_key_provider.NewPgSshKeySecretStore(connPool, secretProvider)
	externalAppsService := &external_applications.ExternalApplicationsService{DbConn: connPool}

	apiServer := api_server.APIServer{
		HealthCheckService:      healthCheckService,
		AuthService:             authService,
		UserCRUDService:         userService,
		UserSecretsStoreService: secretProvider,
		HostServerProvider:      hostServerProvider,
		SshKeyProvider:          sshKeyProvider,
		ExternalAppsService:     externalAppsService,
		UseSsl:                  userHttps,
		Certificate:             certFile,
		CertKey:                 certKey,
		SwaggerSpec:             swaggerSpec,
	}

	switch {
	case initDevUser:
		userService.UpdateUserPasswordById(uuid.Must(uuid.Parse(os.Getenv("DEV_USER_UUID"))), os.Getenv("DEV_APP_PASS"))
	}

	apiServer.StartAPIServices(&srvport)
}
