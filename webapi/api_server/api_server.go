package api_server

import (
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/internal/cors"
	"github.com/babbage88/go-infra/internal/swaggerui"
	"github.com/babbage88/go-infra/services"
	customlogger "github.com/babbage88/go-infra/utils/logger"
	authapi "github.com/babbage88/go-infra/webapi/authapi"
	userapi "github.com/babbage88/go-infra/webapi/user_api_handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartWebApiServer(healthCheckService *services.HealthCheckService, authService authapi.AuthService, userCRUDService *services.UserCRUDService, swaggerSpec []byte, srvadr *string) error {
	mux := http.NewServeMux()
	mux.Handle("/renew", cors.CORSWithPOST(authapi.AuthMiddleware(authapi.Renewcert_renew())))
	mux.Handle("/login", cors.CORSWithPOST(http.HandlerFunc(authapi.LoginHandleFunc(authService))))
	mux.Handle("/dbhealth", cors.CORSWithGET(http.HandlerFunc(healthCheckService.DbReadHealthCheckHandler())))
	mux.Handle("/token/refresh", cors.CORSWithPOST(http.HandlerFunc(authapi.RefreshAuthTokens(authService))))
	mux.Handle("/create/user", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "CreateUser", userapi.CreateUserHandler(userCRUDService))))
	mux.Handle("/update/userpass", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "AlterUser", userapi.UpdateUserPasswordHandler(userCRUDService))))
	mux.Handle("/user/enable", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "AlterUser", userapi.EnableUserHandler(userCRUDService))))
	mux.Handle("/user/disable", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "AlterUser", userapi.DisableUserHandler(userCRUDService))))
	mux.Handle("/user/delete", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "DeleteUser", userapi.SoftDeleteUserHandler(userCRUDService))))
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
	mux.Handle("/authhealthCheck", cors.CORSWithGET(authapi.AuthMiddleware(http.HandlerFunc(authapi.HealthCheckHandler))))
	mux.Handle("/metrics", promhttp.Handler())

	// Add Swagger UI handler
	mux.Handle("/swaggerui/", http.StripPrefix("/swaggerui", swaggerui.ServeSwaggerUI(swaggerSpec)))

	config := customlogger.NewCustomLogger()
	clog := customlogger.SetupLogger(config)

	clog.Info("Starting http server.")
	err := http.ListenAndServe(*srvadr, cors.HandleCORSPreflightMiddleware(mux))
	if err != nil {
		slog.Error("Failed to start server", slog.String("Error", err.Error()))
	}
	return err
}
