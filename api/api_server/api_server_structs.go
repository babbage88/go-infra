package api_server

import (
	authapi "github.com/babbage88/go-infra/api/authapi"
	"github.com/babbage88/go-infra/services/user_crud_svc"
	"github.com/babbage88/go-infra/services/user_secrets"
)

type APIServer struct {
	HealthCheckService      *user_crud_svc.HealthCheckService `json:"apiHealthCheckService"`
	AuthService             authapi.AuthService               `json:"apiAuthService"`
	UserCRUDService         *user_crud_svc.UserCRUDService    `json:"apiUserCrudService"`
	UserSecretsStoreService user_secrets.UserSecretProvider   `json:"apiUserSecretStoreService"`
	SwaggerSpec             []byte                            `json:"apiSwaggerSpec"`
	UseSsl                  bool                              `json:"userSsl"`
	Certificate             string                            `json:"apiCertificate"`
	CertKey                 string                            `json:"apiCertKey"`
}
