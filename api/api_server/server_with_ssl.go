package api_server

import (
	"log/slog"
	"net/http"

	authapi "github.com/babbage88/go-infra/api/authapi"
	userapi "github.com/babbage88/go-infra/api/user_api_handlers"
	"github.com/babbage88/go-infra/internal/cors"
	"github.com/babbage88/go-infra/internal/swaggerui"
	"github.com/babbage88/go-infra/services/user_crud_svc"
	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/babbage88/go-infra/webutils/cert_renew"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartApiServerWithHttps(healthCheckService *user_crud_svc.HealthCheckService,
	authService authapi.AuthService,
	userCRUDService *user_crud_svc.UserCRUDService,
	userSecretStore user_secrets.UserSecretProvider,
	swaggerSpec []byte, srvadr *string,
	certificatePath string,
	certificateKeyPath string,
) error {
	mux := http.NewServeMux()
	mux.Handle("/renew", cors.CORSWithPOST(authapi.AuthMiddleware(cert_renew.Renewcert_renew())))
	mux.Handle("/login", cors.CORSWithPOST(http.HandlerFunc(authapi.LoginHandleFunc(authService))))
	mux.Handle("/dbhealth", cors.CORSWithGET(http.HandlerFunc(healthCheckService.DbReadHealthCheckHandler())))
	mux.Handle("/token/refresh", cors.CORSWithPOST(http.HandlerFunc(authapi.RefreshAuthTokens(authService))))
	mux.Handle("/create/user", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "CreateUser", userapi.CreateUserHandler(userCRUDService))))
	mux.Handle("/update/userpass", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "AlterUser", userapi.UpdateUserPasswordHandler(userCRUDService))))
	mux.Handle("/user/enable", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "AlterUser", userapi.EnableUserHandler(userCRUDService))))
	mux.Handle("/user/disable", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "AlterUser", userapi.DisableUserHandler(userCRUDService))))
	mux.Handle("/user/delete", cors.CORSWithDELETE(authapi.AuthMiddlewareRequirePermission(authService, "DeleteUser", userapi.SoftDeleteUserHandler(userCRUDService))))
	mux.Handle("/user/role", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "AlterUser", userapi.UpdateUserRoleMappingHandler(userCRUDService))))
	mux.Handle("/user/role/remove", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "AlterUser", userapi.DisableUserRoleMappingHandler(userCRUDService))))
	mux.Handle("/create/role", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "CreateRole", userapi.CreateUserRoleHandler(userCRUDService))))
	mux.Handle("/create/permission", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "CreatePermission", userapi.CreateAppPermissionHandler(userCRUDService))))
	mux.Handle("/roles/permission", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "AlterRole", userapi.CreateRolePermissionMappingHandler(userCRUDService))))
	mux.Handle("/roles", cors.CORSWithGET(authapi.AuthMiddlewareRequirePermission(authService, "ReadRoles", userapi.GetAllRolesHandler(userCRUDService))))
	mux.Handle("/users/{ID}", cors.CORSWithGET(authapi.AuthMiddlewareRequirePermission(authService, "ReadUsers", userapi.GetUserByIdHandler(userCRUDService))))
	mux.Handle("/permissions", cors.CORSWithGET(authapi.AuthMiddlewareRequirePermission(authService, "ReadPermissions", userapi.GetAllAppPermissionsHandler(userCRUDService))))
	mux.Handle("/users", cors.CORSWithGET(authapi.AuthMiddleware(userapi.GetAllUsersHandler(userCRUDService))))
	mux.Handle("/healthCheck", cors.CORSWithGET(http.HandlerFunc(authapi.HealthCheckHandler)))
	mux.Handle("/secrets/create/{ID}", cors.CORSWithPOST(user_secrets.CreateSecretHandler(userSecretStore)))
	mux.Handle("/secrets/{ID}", cors.CORSWithGET(user_secrets.GetSecretHandler(userSecretStore)))
	mux.Handle("/user/secrets/{USERID}", cors.CORSWithGET(user_secrets.GetUserSecretEntriesByIdHandler(userSecretStore)))
	mux.Handle("/user/{APPID}/secrets/{USERID}", cors.CORSWithGET(user_secrets.GetUserSecretEntriesByAppIdHandler(userSecretStore)))
	mux.Handle("/secrets/delete/{ID}", cors.CORSWithDELETE(user_secrets.DeleteSecretHandler(userSecretStore)))
	mux.Handle("/authhealthCheck", cors.CORSWithGET(authapi.AuthMiddleware(http.HandlerFunc(authapi.HealthCheckHandler))))
	mux.Handle("/metrics", promhttp.Handler())

	// Add Swagger UI handler
	mux.Handle("/swaggerui/", http.StripPrefix("/swaggerui", swaggerui.ServeSwaggerUI(swaggerSpec)))

	slog.Info("Starting http server.")
	err := http.ListenAndServeTLS(*srvadr, certificatePath, certificateKeyPath, cors.HandleCORSPreflightMiddleware(mux))
	if err != nil {
		slog.Error("Failed to start server", slog.String("Error", err.Error()))
	}
	return err
}
