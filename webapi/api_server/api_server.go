package api_server

import (
	"log/slog"
	"net/http"

	"github.com/babbage88/go-infra/internal/swaggerui"
	"github.com/babbage88/go-infra/services"
	customlogger "github.com/babbage88/go-infra/utils/logger"
	authapi "github.com/babbage88/go-infra/webapi/authapi"
	userapi "github.com/babbage88/go-infra/webapi/user_api_handlers"
	"github.com/babbage88/go-infra/webutils/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartWebApiServer(authService *authapi.UserAuthService, userCRUDService *services.UserCRUDService, swaggerSpec []byte, srvadr *string) error {
	envars := authService.Envars
	mux := http.NewServeMux()
	mux.Handle("/renew", cors.CORSWithPOST(authapi.AuthMiddleware(envars, authapi.Renewcert_renew(envars))))
	mux.Handle("/login", cors.CORSWithPOST(http.HandlerFunc(authapi.LoginHandler(authService))))
	mux.Handle("/token/refresh", cors.CORSWithPOST(http.HandlerFunc(authapi.RefreshAuthTokens(authService))))
	mux.Handle("/create/user", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "CreateUser", userapi.CreateUser(userCRUDService))))
	mux.Handle("/update/userpass", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "AlterUser", userapi.UpdateUserPassword(userCRUDService))))
	mux.Handle("/users", cors.CORSWithPOST(authapi.AuthMiddleware(envars, userapi.GetAllUsers(userCRUDService))))
	mux.Handle("/healthCheck", cors.CORSWithGET(http.HandlerFunc(authapi.HealthCheckHandler)))
	mux.Handle("/authhealthCheck", cors.CORSWithGET(authapi.AuthMiddleware(envars, http.HandlerFunc(authapi.HealthCheckHandler))))
	mux.Handle("/metrics", promhttp.Handler())

	// Add Swagger UI handler
	mux.Handle("/swaggerui/", http.StripPrefix("/swaggerui", swaggerui.ServeSwaggerUI(swaggerSpec)))

	config := customlogger.NewCustomLogger()
	clog := customlogger.SetupLogger(config)

	clog.Info("Starting http server.")
	err := http.ListenAndServe(*srvadr, mux)
	if err != nil {
		slog.Error("Failed to start server", slog.String("Error", err.Error()))
	}
	return err
}
