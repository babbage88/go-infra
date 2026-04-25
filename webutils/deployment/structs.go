package deployment

import coredeploy "github.com/babbage88/infra-core/deployment"

// swagger:model SSHOptions
type SSHOptions = coredeploy.SSHOptions

// swagger:model ProxyInstallRequest
type ProxyInstallRequest = coredeploy.ProxyInstallRequest
// swagger:model ProxyInstallResult
type ProxyInstallResult = coredeploy.ProxyInstallResult
// swagger:model GarageTokenRequest
type GarageTokenRequest = coredeploy.GarageTokenRequest
// swagger:model GarageTokenResult
type GarageTokenResult = coredeploy.GarageTokenResult
// swagger:model GarageNodeRequest
type GarageNodeRequest = coredeploy.GarageNodeRequest
// swagger:model GarageNodeResult
type GarageNodeResult = coredeploy.GarageNodeResult
// swagger:model ValkeyInstallRequest
type ValkeyInstallRequest = coredeploy.ValkeyInstallRequest
// swagger:model ValkeyInstallResult
type ValkeyInstallResult = coredeploy.ValkeyInstallResult
// swagger:model MariaDBInstallRequest
type MariaDBInstallRequest = coredeploy.MariaDBInstallRequest
// swagger:model MariaDBInstallResult
type MariaDBInstallResult = coredeploy.MariaDBInstallResult
// swagger:model SystemdAppDeployRequest
type SystemdAppDeployRequest = coredeploy.SystemdAppDeployRequest
// swagger:model SystemdAppDeployResult
type SystemdAppDeployResult = coredeploy.SystemdAppDeployResult
// swagger:model PostgresAppSetupRequest
type PostgresAppSetupRequest = coredeploy.PostgresAppSetupRequest
// swagger:model PostgresAppSetupResult
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
