package deployment

import coredeploy "github.com/babbage88/infra-core/deployment"

type SSHOptions = coredeploy.SSHOptions

type ProxyInstallRequest = coredeploy.ProxyInstallRequest
type ProxyInstallResult = coredeploy.ProxyInstallResult
type GarageTokenRequest = coredeploy.GarageTokenRequest
type GarageTokenResult = coredeploy.GarageTokenResult
type GarageNodeRequest = coredeploy.GarageNodeRequest
type GarageNodeResult = coredeploy.GarageNodeResult
type ValkeyInstallRequest = coredeploy.ValkeyInstallRequest
type ValkeyInstallResult = coredeploy.ValkeyInstallResult
type MariaDBInstallRequest = coredeploy.MariaDBInstallRequest
type MariaDBInstallResult = coredeploy.MariaDBInstallResult
type SystemdAppDeployRequest = coredeploy.SystemdAppDeployRequest
type SystemdAppDeployResult = coredeploy.SystemdAppDeployResult
type PostgresAppSetupRequest = coredeploy.PostgresAppSetupRequest
type PostgresAppSetupResult = coredeploy.PostgresAppSetupResult

// swagger:parameters InstallProxy
type InstallProxyParams struct {
	// Proxy name to install.
	// in: path
	// required: true
	Name string `json:"name"`
	// in: body
	Body ProxyInstallRequest
}

// swagger:parameters InstallMariaDB
type InstallMariaDBParams struct {
	// in: body
	Body MariaDBInstallRequest
}

// swagger:parameters InstallValkey
type InstallValkeyParams struct {
	// in: body
	Body ValkeyInstallRequest
}

// swagger:parameters DeployGarageNode
type DeployGarageNodeParams struct {
	// in: body
	Body GarageNodeRequest
}

// swagger:parameters CreateGarageToken
type CreateGarageTokenParams struct {
	// in: body
	Body GarageTokenRequest
}

// swagger:parameters DeploySystemdApp
type DeploySystemdAppParams struct {
	// in: body
	Body SystemdAppDeployRequest
}

// swagger:parameters SetupPostgresApp
type SetupPostgresAppParams struct {
	// in: body
	Body PostgresAppSetupRequest
}

// swagger:response ProxyInstallResponse
type ProxyInstallResponse struct {
	// in: body
	Body ProxyInstallResult
}

// swagger:response MariaDBInstallResponse
type MariaDBInstallResponse struct {
	// in: body
	Body MariaDBInstallResult
}

// swagger:response ValkeyInstallResponse
type ValkeyInstallResponse struct {
	// in: body
	Body ValkeyInstallResult
}

// swagger:response GarageNodeResponse
type GarageNodeResponse struct {
	// in: body
	Body GarageNodeResult
}

// swagger:response GarageTokenResponse
type GarageTokenResponse struct {
	// in: body
	Body GarageTokenResult
}

// swagger:response SystemdAppDeployResponse
type SystemdAppDeployResponse struct {
	// in: body
	Body SystemdAppDeployResult
}

// swagger:response PostgresAppSetupResponse
type PostgresAppSetupResponse struct {
	// in: body
	Body PostgresAppSetupResult
}
