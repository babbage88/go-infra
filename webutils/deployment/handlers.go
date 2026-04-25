package deployment

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	coredeploy "github.com/babbage88/infra-core/deployment"
)

// swagger:route POST /api/v1/proxy/{name}/install Deployment InstallProxy
// Install and configure a supported proxy on a remote host over SSH.
// responses:
//
//	200: ProxyInstallResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func InstallProxyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxyName := r.PathValue("name")
		defaultReq, err := coredeploy.DefaultProxyInstallRequest(proxyName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		req := coredeploy.ProxyInstallRequest{}
		if !decodeJSONBody(w, r, &req) {
			return
		}
		req = coredeploy.MergeProxyInstallDefaults(req, defaultReq)

		result, err := coredeploy.InstallProxy(req)
		if err != nil {
			slog.Error("failed to install proxy", slog.String("proxy", proxyName), slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route POST /api/v1/database/mariadb/install Deployment InstallMariaDB
// Install and configure MariaDB for remote access.
// responses:
//
//	200: MariaDBInstallResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func InstallMariaDBHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := coredeploy.MariaDBInstallRequest{}
		if !decodeJSONBody(w, r, &req) {
			return
		}
		req = coredeploy.MergeMariaDBInstallDefaults(req, coredeploy.DefaultMariaDBInstallRequest())

		result, err := coredeploy.InstallMariaDB(req)
		if err != nil {
			slog.Error("failed to install mariadb", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route POST /api/v1/database/valkey/install Deployment InstallValkey
// Install and configure Valkey for remote access.
// responses:
//
//	200: ValkeyInstallResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func InstallValkeyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := coredeploy.ValkeyInstallRequest{}
		if !decodeJSONBody(w, r, &req) {
			return
		}
		req = coredeploy.MergeValkeyInstallDefaults(req, coredeploy.DefaultValkeyInstallRequest())

		result, err := coredeploy.InstallValkey(req)
		if err != nil {
			slog.Error("failed to install valkey", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route POST /api/v1/storage/s3/garage/node Deployment DeployGarageNode
// Install and configure a Garage S3 node on a remote host.
// responses:
//
//	200: GarageNodeResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func DeployGarageNodeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := coredeploy.GarageNodeRequest{}
		if !decodeJSONBody(w, r, &req) {
			return
		}
		req = coredeploy.MergeGarageNodeDefaults(req, coredeploy.DefaultGarageNodeRequest())

		result, err := coredeploy.DeployGarageNode(req)
		if err != nil {
			slog.Error("failed to deploy garage node", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

// swagger:route POST /api/v1/storage/s3/garage/token Deployment CreateGarageToken
// Create Garage S3 credentials on a remote host.
// responses:
//
//	200: GarageTokenResponse
//	400: description:Invalid request
//	401: description:Unauthorized
//	500: description:Internal Server Error
func CreateGarageTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := coredeploy.GarageTokenRequest{}
		if !decodeJSONBody(w, r, &req) {
			return
		}
		req = coredeploy.MergeGarageTokenDefaults(req, coredeploy.DefaultGarageTokenRequest())

		result, err := coredeploy.CreateGarageToken(req)
		if err != nil {
			slog.Error("failed to create garage token", slog.String("error", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, result)
	}
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dest interface{}) bool {
	if r.Body == nil {
		return true
	}
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
