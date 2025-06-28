package api_server

import (
	"log/slog"
	"net/http"

	authapi "github.com/babbage88/go-infra/api/authapi"
	userapi "github.com/babbage88/go-infra/api/user_api_handlers"
	"github.com/babbage88/go-infra/internal/cors"
	"github.com/babbage88/go-infra/internal/swaggerui"
	"github.com/babbage88/go-infra/services/external_applications"
	"github.com/babbage88/go-infra/services/host_servers"
	"github.com/babbage88/go-infra/services/node_networking"
	"github.com/babbage88/go-infra/services/ssh_connections"
	"github.com/babbage88/go-infra/services/ssh_key_provider"
	"github.com/babbage88/go-infra/services/user_crud_svc"
	"github.com/babbage88/go-infra/services/user_secrets"
	"github.com/babbage88/go-infra/webutils/cert_renew"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// hostServerByIDHandler handles GET, PUT, DELETE methods for /host-servers/{ID}
func hostServerByIDHandler(provider host_servers.HostServerProvider, authService authapi.AuthService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Read permission
			authapi.AuthMiddlewareRequirePermission(authService, "ReadHostServers", host_servers.GetHostServerHandler(provider)).ServeHTTP(w, r)
		case http.MethodPut:
			// Manage permission
			authapi.AuthMiddlewareRequirePermission(authService, "ManageHostServers", host_servers.UpdateHostServerHandler(provider)).ServeHTTP(w, r)
		case http.MethodDelete:
			// Manage permission
			authapi.AuthMiddlewareRequirePermission(authService, "ManageHostServers", host_servers.DeleteHostServerHandler(provider)).ServeHTTP(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
}

func AddApplicationRoutes(mux *http.ServeMux, healthCheckService *user_crud_svc.HealthCheckService, authService authapi.AuthService, userCRUDService *user_crud_svc.UserCRUDService,
	userSecretStore user_secrets.UserSecretProvider, hostServerProvider host_servers.HostServerProvider, sshKeyProvider ssh_key_provider.SshKeySecretProvider, externalAppsService external_applications.ExternalApplications, swaggerSpec []byte, sshConnectionManager *ssh_connections.SSHConnectionManager) {
	mux.Handle("/renew", cors.CORSWithPOST(authapi.AuthMiddleware(cert_renew.Renewcert_renew())))
	mux.Handle("/login", cors.CORSWithPOST(authapi.LoginHandler(authService)))
	mux.Handle("/dbhealth", cors.CORSWithGET(healthCheckService.DbReadHealthCheckHandler()))
	mux.Handle("/token/verify", cors.CORSWithPOST(authapi.VerifyTokenHandler(authService)))
	mux.Handle("/token/refresh", cors.CORSWithPOST(authapi.RefreshAccessTokensHandler(authService)))
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
	mux.Handle("/secrets/create", cors.CORSWithPOST(user_secrets.CreateSecretHandler(userSecretStore)))
	mux.Handle("/secrets/{ID}", cors.CORSWithGET(user_secrets.GetSecretHandler(userSecretStore)))
	mux.Handle("/user/secrets/{USERID}", cors.CORSWithGET(user_secrets.GetUserSecretEntriesByIdHandler(userSecretStore)))
	mux.Handle("/user/{APPID}/secrets/{USERID}", cors.CORSWithGET(user_secrets.GetUserSecretEntriesByAppIdHandler(userSecretStore)))
	mux.Handle("/user/secrets/by-name/{APPNAME}/{USERID}", cors.CORSWithGET(user_secrets.GetUserSecretEntriesByAppNameHandler(userSecretStore)))
	mux.Handle("/secrets/delete/{ID}", cors.CORSWithDELETE(user_secrets.DeleteSecretHandler(userSecretStore)))
	mux.Handle("/authhealthCheck", cors.CORSWithGET(authapi.AuthMiddleware(http.HandlerFunc(authapi.HealthCheckHandler))))
	mux.Handle("/metrics", promhttp.Handler())

	// Host server routes
	mux.Handle("/host-servers/create", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "ManageHostServers", host_servers.CreateHostServerHandler(hostServerProvider))))
	mux.Handle("/host-servers/{ID}", cors.CORSWithMethods(
		hostServerByIDHandler(hostServerProvider, authService),
		http.MethodGet, http.MethodPut, http.MethodDelete,
	))
	mux.Handle("/host-servers", cors.CORSWithGET(authapi.AuthMiddlewareRequirePermission(authService, "ReadHostServers", host_servers.GetAllHostServersHandler(hostServerProvider))))

	// SSH key routes
	mux.Handle("/ssh-keys/create", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "ManageSshKeys", ssh_key_provider.CreateSshKeyHandler(sshKeyProvider))))
	mux.Handle("/ssh-keys/{id}", cors.CORSWithDELETE(authapi.AuthMiddlewareRequirePermission(authService, "ManageSshKeys", ssh_key_provider.DeleteSshKeyHandler(sshKeyProvider))))
	mux.Handle("/ssh-keys/user/{userId}", cors.CORSWithGET(authapi.AuthMiddlewareRequirePermission(authService, "ReadSshKeys", ssh_key_provider.GetSshKeysByUserIdHandler(sshKeyProvider))))

	// SSH key host mapping routes
	mux.Handle("/ssh-key-host-mappings/create", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "ManageSshKeys", ssh_key_provider.CreateSshKeyHostMappingHandler(sshKeyProvider))))
	mux.Handle("/ssh-key-host-mappings/{id}", cors.CORSWithMethods(
		ssh_key_provider.SshKeyHostMappingByIDHandler(sshKeyProvider, authService),
		http.MethodGet, http.MethodPut, http.MethodDelete,
	))
	mux.Handle("/ssh-key-host-mappings/user/{userId}", cors.CORSWithGET(authapi.AuthMiddlewareRequirePermission(authService, "ReadSshKeys", ssh_key_provider.GetSshKeyHostMappingsByUserIdHandler(sshKeyProvider))))
	mux.Handle("/ssh-key-host-mappings/host/{hostId}", cors.CORSWithGET(authapi.AuthMiddlewareRequirePermission(authService, "ReadSshKeys", ssh_key_provider.GetSshKeyHostMappingsByHostIdHandler(sshKeyProvider))))
	mux.Handle("/ssh-key-host-mappings/key/{keyId}", cors.CORSWithGET(authapi.AuthMiddlewareRequirePermission(authService, "ReadSshKeys", ssh_key_provider.GetSshKeyHostMappingsByKeyIdHandler(sshKeyProvider))))

	// SSH connection routes
	if sshConnectionManager != nil {
		mux.Handle("POST /ssh/connect", cors.CORSWithPOST(authapi.AuthMiddlewareRequirePermission(authService, "SshConnect", http.HandlerFunc(sshConnectionManager.CreateSSHConnectionHandler))))
		mux.Handle("DELETE /ssh/connect/{CONNID}", cors.CORSWithDELETE(authapi.AuthMiddlewareRequirePermission(authService, "SshConnect", http.HandlerFunc(sshConnectionManager.CloseSSHConnectionHandler))))
		mux.Handle("GET /ssh/websocket/{CONNID}", http.HandlerFunc(sshConnectionManager.SSHWebSocketHandler))
	}

	// External applications routes
	mux.Handle("/external-applications", cors.CORSWithMethods(
		external_applications.ExternalApplicationsHandler(externalAppsService, authService),
		http.MethodGet, http.MethodPost,
	))
	mux.Handle("/external-applications/{ID}", cors.CORSWithMethods(
		external_applications.ExternalApplicationByIDHandler(externalAppsService, authService),
		http.MethodGet, http.MethodPut, http.MethodDelete,
	))
	mux.Handle("/external-applications/by-name/{name}", cors.CORSWithMethods(
		external_applications.ExternalApplicationByNameHandler(externalAppsService, authService),
		http.MethodGet, http.MethodDelete,
	))
	mux.Handle("/external-applications/id/{name}", cors.CORSWithGET(
		external_applications.GetExternalApplicationIdByNameHandler(externalAppsService),
	))
	mux.Handle("/external-applications/name/{ID}", cors.CORSWithGET(
		external_applications.GetExternalApplicationNameByIdHandler(externalAppsService),
	))

	// Add Swagger UI handler
	mux.Handle("/swaggerui/", http.StripPrefix("/swaggerui", swaggerui.ServeSwaggerUI(swaggerSpec)))

	// Register network pinger routes
	SetupNetworkPingerRoutes(mux, hostServerProvider, authService)
}

// SetupNetworkPingerRoutes sets up all the network pinger routes
func SetupNetworkPingerRoutes(
	router *http.ServeMux,
	hostServerProvider host_servers.HostServerProvider,
	authService authapi.AuthService,
) {
	pinger := node_networking.NewNetworkPinger(hostServerProvider)

	router.Handle("POST /network/ping",
		authapi.AuthMiddlewareRequirePermission(authService, "NetworkPing", node_networking.PingHandler(pinger)))

	router.Handle("POST /network/ping-host-server",
		authapi.AuthMiddlewareRequirePermission(authService, "NetworkPing", node_networking.PingHostServerHandler(pinger)))

	router.Handle("POST /network/probe-tcp-hostname",
		authapi.AuthMiddlewareRequirePermission(authService, "NetworkProbe", node_networking.ProbeTCPByHostnameHandler(pinger)))

	router.Handle("POST /network/probe-udp-hostname",
		authapi.AuthMiddlewareRequirePermission(authService, "NetworkProbe", node_networking.ProbeUDPByHostnameHandler(pinger)))

	router.Handle("POST /network/probe-tcp-host-id",
		authapi.AuthMiddlewareRequirePermission(authService, "NetworkProbe", node_networking.ProbeTCPByHostIdHandler(pinger)))

	router.Handle("POST /network/probe-udp-host-id",
		authapi.AuthMiddlewareRequirePermission(authService, "NetworkProbe", node_networking.ProbeUDPByHostIdHandler(pinger)))

	router.Handle("GET /network/ping/{target}",
		authapi.AuthMiddlewareRequirePermission(authService, "NetworkPing", node_networking.PingGetHandler(pinger)))

	router.Handle("GET /network/probe-tcp/{target}/{port}",
		authapi.AuthMiddlewareRequirePermission(authService, "NetworkProbe", node_networking.ProbeTCPGetHandler(pinger)))

	router.Handle("GET /network/probe-udp/{target}/{port}",
		authapi.AuthMiddlewareRequirePermission(authService, "NetworkPing", node_networking.ProbeUDPGetHandler(pinger)))
}

func (api *APIServer) StartAPIServices(srvadr *string) error {
	mux := http.NewServeMux()
	AddApplicationRoutes(mux, api.HealthCheckService, api.AuthService, api.UserCRUDService, api.UserSecretsStoreService, api.HostServerProvider, api.SshKeyProvider, api.ExternalAppsService, api.SwaggerSpec, api.SSHConnectionManager)

	// Start a dedicated WebSocket server on :8090 with no middleware for /ssh/websocket/{connectionId}
	go func() {
		wsMux := http.NewServeMux()
		if api.SSHConnectionManager != nil {
			wsMux.Handle("/ssh/websocket/", http.HandlerFunc(api.SSHConnectionManager.SSHWebSocketHandler))
		}
		slog.Info("Starting dedicated WebSocket server on :8090 (no middleware)")
		http.ListenAndServe(":8090", wsMux)
	}()

	switch {
	case api.UseSsl:
		slog.Info("Starting https server.", slog.String("ListenAddress", *srvadr))
		err := http.ListenAndServeTLS(*srvadr, api.Certificate, api.CertKey, requestLoggingMiddleware(cors.HandleCORSPreflightMiddleware(mux)))

		if err != nil {
			slog.Error("Failed to start server", slog.String("Error", err.Error()))
		}
		return err
	default:
		slog.Info("Starting http server.", slog.String("ListenAddress", *srvadr))
		err := http.ListenAndServe(*srvadr, requestLoggingMiddleware(cors.HandleCORSPreflightMiddleware(mux)))
		if err != nil {
			slog.Error("Failed to start server", slog.String("Error", err.Error()))
		}
		return err
	}
}
