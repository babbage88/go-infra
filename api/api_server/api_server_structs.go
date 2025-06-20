package api_server

import (
	authapi "github.com/babbage88/go-infra/api/authapi"
	"github.com/babbage88/go-infra/services/external_applications"
	"github.com/babbage88/go-infra/services/host_servers"
	"github.com/babbage88/go-infra/services/ssh_key_provider"
	"github.com/babbage88/go-infra/services/user_crud_svc"
	"github.com/babbage88/go-infra/services/user_secrets"
)

// APIServer represents the API server configuration
type APIServer struct {
	HealthCheckService      *user_crud_svc.HealthCheckService
	AuthService             authapi.AuthService
	UserCRUDService         *user_crud_svc.UserCRUDService
	UserSecretsStoreService user_secrets.UserSecretProvider
	HostServerProvider      host_servers.HostServerProvider
	SshKeyProvider          ssh_key_provider.SshKeySecretProvider
	ExternalAppsService     external_applications.ExternalApplications
	UseSsl                  bool
	Certificate             string
	CertKey                 string
	SwaggerSpec             []byte
}
