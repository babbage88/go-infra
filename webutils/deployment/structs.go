package deployment

// swagger:model SSHOptions
type SSHOptions struct {
	Host             string `json:"host,omitempty"`
	User             string `json:"user,omitempty"`
	KeyPath          string `json:"key_path,omitempty"`
	PrivateKeyPEM    string `json:"private_key_pem,omitempty"`
	PrivateKeyBase64 string `json:"private_key_base64,omitempty"`
	Passphrase       string `json:"passphrase,omitempty"`
	UseAgent         bool   `json:"use_agent,omitempty"`
	Port             uint   `json:"port,omitempty"`
}

// swagger:model ProxyInstallRequest
type ProxyInstallRequest struct {
	SSH             SSHOptions `json:"ssh,omitempty"`
	Name            string     `json:"name,omitempty"`
	PackageName     string     `json:"package_name,omitempty"`
	BinaryName      string     `json:"binary_name,omitempty"`
	ServiceName     string     `json:"service_name,omitempty"`
	ConfigPath      string     `json:"config_path,omitempty"`
	LocalConfigPath string     `json:"local_config_path,omitempty"`
}

// swagger:model ProxyInstallResult
type ProxyInstallResult struct {
	Host            string `json:"host"`
	Name            string `json:"name"`
	PackageName     string `json:"package_name"`
	BinaryName      string `json:"binary_name"`
	ServiceName     string `json:"service_name"`
	ConfigPath      string `json:"config_path"`
	LocalConfigPath string `json:"local_config_path,omitempty"`
}

// swagger:model GarageTokenRequest
type GarageTokenRequest struct {
	SSH                SSHOptions `json:"ssh,omitempty"`
	BucketName         string     `json:"bucket_name,omitempty"`
	KeyName            string     `json:"key_name,omitempty"`
	CreateBucket       *bool      `json:"create_bucket,omitempty"`
	AllowCreateBuckets *bool      `json:"allow_create_buckets,omitempty"`
	AllowRead          *bool      `json:"allow_read,omitempty"`
	AllowWrite         *bool      `json:"allow_write,omitempty"`
	AllowOwner         *bool      `json:"allow_owner,omitempty"`
	BinaryPath         string     `json:"binary_path,omitempty"`
	ConfigPath         string     `json:"config_path,omitempty"`
	S3Endpoint         string     `json:"s3_endpoint,omitempty"`
	LayoutZone         string     `json:"layout_zone,omitempty"`
	LayoutCapacity     string     `json:"layout_capacity,omitempty"`
}

// swagger:model GarageTokenResult
type GarageTokenResult struct {
	Host              string `json:"host"`
	S3Endpoint        string `json:"s3_endpoint"`
	BucketName        string `json:"bucket_name,omitempty"`
	KeyName           string `json:"key_name"`
	AccessKeyID       string `json:"access_key_id"`
	SecretAccessKey   string `json:"secret_access_key"`
	MCAliasSetCommand string `json:"mc_alias_set_command"`
}

// swagger:model GarageNodeRequest
type GarageNodeRequest struct {
	SSH               SSHOptions `json:"ssh,omitempty"`
	Version           string     `json:"version,omitempty"`
	BinaryPath        string     `json:"binary_path,omitempty"`
	ConfigPath        string     `json:"config_path,omitempty"`
	MetadataDir       string     `json:"metadata_dir,omitempty"`
	DataDir           string     `json:"data_dir,omitempty"`
	DBEngine          string     `json:"db_engine,omitempty"`
	ReplicationFactor int        `json:"replication_factor,omitempty"`
	RPCBindAddr       string     `json:"rpc_bind_addr,omitempty"`
	RPCPublicAddr     string     `json:"rpc_public_addr,omitempty"`
	RPCSecret         string     `json:"rpc_secret,omitempty"`
	S3APIBindAddr     string     `json:"s3_api_bind_addr,omitempty"`
	S3Region          string     `json:"s3_region,omitempty"`
	S3RootDomain      string     `json:"s3_root_domain,omitempty"`
	S3WebBindAddr     string     `json:"s3_web_bind_addr,omitempty"`
	S3WebRootDomain   string     `json:"s3_web_root_domain,omitempty"`
	S3WebIndex        string     `json:"s3_web_index,omitempty"`
	K2VAPIBindAddr    string     `json:"k2v_api_bind_addr,omitempty"`
	AdminAPIBindAddr  string     `json:"admin_api_bind_addr,omitempty"`
	AdminToken        string     `json:"admin_token,omitempty"`
	MetricsToken      string     `json:"metrics_token,omitempty"`
	LogLevel          string     `json:"log_level,omitempty"`
}

// swagger:model GarageNodeResult
type GarageNodeResult struct {
	Host          string `json:"host"`
	BinaryPath    string `json:"binary_path"`
	ConfigPath    string `json:"config_path"`
	ServiceName   string `json:"service_name"`
	RPCPublicAddr string `json:"rpc_public_addr"`
	S3Endpoint    string `json:"s3_endpoint"`
	AdminEndpoint string `json:"admin_endpoint"`
	AdminToken    string `json:"admin_token"`
	MetricsToken  string `json:"metrics_token"`
}

// swagger:model ValkeyInstallRequest
type ValkeyInstallRequest struct {
	SSH      SSHOptions `json:"ssh,omitempty"`
	Username string     `json:"username,omitempty"`
	Password string     `json:"password,omitempty"`
	Bind     string     `json:"bind,omitempty"`
	Port     int        `json:"port,omitempty"`
	ACLFile  string     `json:"acl_file,omitempty"`
}

// swagger:model ValkeyInstallResult
type ValkeyInstallResult struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	URI      string `json:"uri"`
}

// swagger:model MariaDBInstallRequest
type MariaDBInstallRequest struct {
	SSH          SSHOptions `json:"ssh,omitempty"`
	DatabaseName string     `json:"db_name,omitempty"`
	Username     string     `json:"username,omitempty"`
	Password     string     `json:"password,omitempty"`
	Bind         string     `json:"bind,omitempty"`
	Port         int        `json:"port,omitempty"`
}

// swagger:model MariaDBInstallResult
type MariaDBInstallResult struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DatabaseName string `json:"db_name"`
	Username     string `json:"username"`
	URI          string `json:"uri"`
}

// swagger:model SystemdAppDeployRequest
type SystemdAppDeployRequest struct {
	SSH               SSHOptions        `json:"ssh,omitempty"`
	AppName           string            `json:"app_name,omitempty"`
	EnvVars           map[string]string `json:"env_vars,omitempty"`
	ServiceUser       string            `json:"service_user,omitempty"`
	ServiceUID        int64             `json:"service_uid,omitempty"`
	DestinationBinary string            `json:"destination_binary,omitempty"`
	InstallDir        string            `json:"install_dir,omitempty"`
	SystemdDir        string            `json:"systemd_dir,omitempty"`
	SourceDir         string            `json:"source_dir,omitempty"`
	SourceBin         string            `json:"source_bin,omitempty"`
	SourceGoModule    string            `json:"source_go_module,omitempty"`
	SourceRepo        string            `json:"source_repo,omitempty"`
	SourceRef         string            `json:"source_ref,omitempty"`
	SourcePackage     string            `json:"source_package,omitempty"`
	SourceExcludes    []string          `json:"source_excludes,omitempty"`
}

// swagger:model SystemdAppDeployResult
type SystemdAppDeployResult struct {
	Host         string `json:"host"`
	AppName      string `json:"app_name"`
	ServiceName  string `json:"service_name"`
	ServiceUser  string `json:"service_user"`
	InstallDir   string `json:"install_dir"`
	BinaryPath   string `json:"binary_path"`
	SystemdUnit  string `json:"systemd_unit"`
	SourceBinary string `json:"source_binary,omitempty"`
}

// swagger:model PostgresAppSetupRequest
type PostgresAppSetupRequest struct {
	SSH                           SSHOptions `json:"ssh,omitempty"`
	DatabaseName                  string     `json:"db_name,omitempty"`
	Username                      string     `json:"username,omitempty"`
	Password                      string     `json:"password,omitempty"`
	SchemaName                    string     `json:"schema_name,omitempty"`
	CreateDB                      *bool      `json:"create_db,omitempty"`
	DropFirst                     *bool      `json:"drop_first,omitempty"`
	PostgresUser                  string     `json:"postgres_user,omitempty"`
	PostgresPassword              string     `json:"postgres_password,omitempty"`
	PostgresHost                  string     `json:"postgres_host,omitempty"`
	PostgresPort                  int        `json:"postgres_port,omitempty"`
	PostgresConnDB                string     `json:"postgres_conn_db,omitempty"`
	SetupRemotePostgres           *bool      `json:"setup_remote_postgres,omitempty"`
	RemotePostgresHBACIDR         string     `json:"remote_postgres_hba_cidr,omitempty"`
	RemotePostgresAuthMethod      string     `json:"remote_postgres_auth_method,omitempty"`
	RemotePostgresListenAddresses string     `json:"remote_postgres_listen_addresses,omitempty"`
}

// swagger:model PostgresAppSetupResult
type PostgresAppSetupResult struct {
	Host         string `json:"host"`
	DatabaseName string `json:"db_name"`
	Username     string `json:"username"`
	SchemaName   string `json:"schema_name"`
	PostgresHost string `json:"postgres_host"`
	PostgresPort int    `json:"postgres_port"`
	URI          string `json:"uri"`
}

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
