package proxmox

import "github.com/google/uuid"

// swagger:model ProxmoxAuthOptions
type ProxmoxAuthOptions struct {
	HostURL    string `json:"host_url,omitempty"`
	APIToken   string `json:"api_token,omitempty"`
	APITokenID string `json:"api_token_id,omitempty"`
	APISecret  string `json:"api_secret,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	UseToken   *bool  `json:"use_token,omitempty"`
	SkipTLS    *bool  `json:"skip_tls,omitempty"`
}

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

// swagger:model ProxmoxVM
type ProxmoxVM struct {
	VMID           int     `json:"vmid"`
	Name           string  `json:"name"`
	Status         string  `json:"status"`
	CPU            float64 `json:"cpu,omitempty"`
	MaxMem         int64   `json:"maxmem,omitempty"`
	Mem            int64   `json:"mem,omitempty"`
	MaxDisk        int64   `json:"maxdisk,omitempty"`
	Node           string  `json:"node,omitempty"`
	Tags           string  `json:"tags,omitempty"`
	Template       int     `json:"template,omitempty"`
	Uptime         int64   `json:"uptime,omitempty"`
	RunningMachine string  `json:"running-machine,omitempty"`
	RunningQemu    string  `json:"running-qemu,omitempty"`
}

// swagger:model ProxmoxVMListResult
type ProxmoxVMListResult struct {
	Node string      `json:"node"`
	VMs  []ProxmoxVM `json:"vms"`
}

// swagger:model ProxmoxContainer
type ProxmoxContainer struct {
	VMID     int     `json:"vmid"`
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	CPU      float64 `json:"cpu,omitempty"`
	MaxMem   int64   `json:"maxmem,omitempty"`
	Mem      int64   `json:"mem,omitempty"`
	MaxDisk  int64   `json:"maxdisk,omitempty"`
	Disk     int64   `json:"disk,omitempty"`
	Node     string  `json:"node,omitempty"`
	Tags     string  `json:"tags,omitempty"`
	Template int     `json:"template,omitempty"`
	Uptime   int64   `json:"uptime,omitempty"`
}

// swagger:model ProxmoxContainerListResult
type ProxmoxContainerListResult struct {
	Node       string             `json:"node"`
	Containers []ProxmoxContainer `json:"containers"`
}

// swagger:model ProxmoxWorkload
type ProxmoxWorkload struct {
	Kind     string  `json:"kind"`
	VMID     int     `json:"vmid"`
	Name     string  `json:"name"`
	Status   string  `json:"status"`
	CPU      float64 `json:"cpu,omitempty"`
	MaxMem   int64   `json:"maxmem,omitempty"`
	Mem      int64   `json:"mem,omitempty"`
	MaxDisk  int64   `json:"maxdisk,omitempty"`
	Disk     int64   `json:"disk,omitempty"`
	Node     string  `json:"node,omitempty"`
	Tags     string  `json:"tags,omitempty"`
	Template int     `json:"template,omitempty"`
	Uptime   int64   `json:"uptime,omitempty"`
}

// swagger:model ProxmoxWorkloadInventoryResult
type ProxmoxWorkloadInventoryResult struct {
	Node           string            `json:"node"`
	VMCount        int               `json:"vmCount"`
	ContainerCount int               `json:"containerCount"`
	WorkloadCount  int               `json:"workloadCount"`
	Workloads      []ProxmoxWorkload `json:"workloads"`
}

// swagger:model ProxmoxVMStartRequest
type ProxmoxVMStartRequest struct {
	Auth            ProxmoxAuthOptions `json:"auth,omitempty"`
	HostServerID    *uuid.UUID         `json:"host_server_id,omitempty"`
	ProxmoxSecretID *uuid.UUID         `json:"proxmox_secret_id,omitempty"`
	Node            string             `json:"node,omitempty"`
	VMID            int                `json:"vmid,omitempty"`
}

// swagger:model ProxmoxVMStartResult
type ProxmoxVMStartResult struct {
	Node string `json:"node"`
	VMID int    `json:"vmid"`
	UPID string `json:"upid,omitempty"`
}

// swagger:model ProxmoxLXCRequest
type ProxmoxLXCRequest struct {
	Auth            ProxmoxAuthOptions `json:"auth,omitempty"`
	HostServerID    *uuid.UUID         `json:"host_server_id,omitempty"`
	ProxmoxSecretID *uuid.UUID         `json:"proxmox_secret_id,omitempty"`
	Node            string             `json:"node,omitempty"`
	VMID            int                `json:"vmid,omitempty"`
	Hostname        string             `json:"hostname,omitempty"`
	Password        string             `json:"password,omitempty"`
	OSTemplate      string             `json:"ostemplate,omitempty"`
	SshPublicKeys   []string           `json:"ssh_public_keys,omitempty"`
	Storage         string             `json:"storage,omitempty"`
	RootFSSize      string             `json:"rootfs_size,omitempty"`
	Memory          int                `json:"memory,omitempty"`
	Swap            int                `json:"swap,omitempty"`
	Cores           int                `json:"cores,omitempty"`
	CPULimit        int                `json:"cpu_limit,omitempty"`
	CPUUnits        int                `json:"cpu_units,omitempty"`
	Net0            string             `json:"net0,omitempty"`
	Arch            string             `json:"arch,omitempty"`
	Cmode           string             `json:"cmode,omitempty"`
	Features        string             `json:"features,omitempty"`
	Nameserver      string             `json:"nameserver,omitempty"`
	SearchDomain    string             `json:"search_domain,omitempty"`
	Description     string             `json:"description,omitempty"`
	Unprivileged    *bool              `json:"unprivileged,omitempty"`
	Start           *bool              `json:"start,omitempty"`
	Console         *bool              `json:"console,omitempty"`
}

// swagger:model ProxmoxLXCResult
type ProxmoxLXCResult struct {
	Node     string `json:"node"`
	VMID     int    `json:"vmid"`
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
	Started  bool   `json:"started"`
	Console  bool   `json:"console"`
}

// swagger:model ProxmoxVMCreateRequest
type ProxmoxVMCreateRequest struct {
	Auth              ProxmoxAuthOptions `json:"auth,omitempty"`
	SSH               SSHOptions         `json:"ssh,omitempty"`
	HostServerID      *uuid.UUID         `json:"host_server_id,omitempty"`
	ProxmoxSecretID   *uuid.UUID         `json:"proxmox_secret_id,omitempty"`
	Node              string             `json:"node,omitempty"`
	VMID              int                `json:"vmid,omitempty"`
	TemplateVMID      int                `json:"template_vmid,omitempty"`
	Name              string             `json:"name,omitempty"`
	MemoryMB          int                `json:"memory_mb,omitempty"`
	Sockets           int                `json:"sockets,omitempty"`
	Cores             int                `json:"cores,omitempty"`
	Description       string             `json:"description,omitempty"`
	Storage           string             `json:"storage,omitempty"`
	FullClone         *bool              `json:"full_clone,omitempty"`
	Start             *bool              `json:"start,omitempty"`
	CIUser            string             `json:"ci_user,omitempty"`
	CIPassword        string             `json:"ci_password,omitempty"`
	SshPublicKeys     []string           `json:"ssh_public_keys,omitempty"`
	IPConfig0         string             `json:"ipconfig0,omitempty"`
	Nameserver        string             `json:"nameserver,omitempty"`
	SearchDomain      string             `json:"search_domain,omitempty"`
	CISnippetsStorage string             `json:"ci_snippets_storage,omitempty"`
	CICustomScript    string             `json:"ci_custom_script,omitempty"`
}

// swagger:model ProxmoxVMCreateResult
type ProxmoxVMCreateResult struct {
	Node         string `json:"node"`
	VMID         int    `json:"vmid"`
	TemplateVMID int    `json:"template_vmid"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Started      bool   `json:"started"`
}

// swagger:model ProxmoxVMTemplateRequest
type ProxmoxVMTemplateRequest struct {
	Auth             ProxmoxAuthOptions `json:"auth,omitempty"`
	SSH              SSHOptions         `json:"ssh,omitempty"`
	HostServerID     *uuid.UUID         `json:"host_server_id,omitempty"`
	ProxmoxSecretID  *uuid.UUID         `json:"proxmox_secret_id,omitempty"`
	Node             string             `json:"node,omitempty"`
	VMID             int                `json:"vmid,omitempty"`
	Name             string             `json:"name,omitempty"`
	ImageURL         string             `json:"image_url,omitempty"`
	Storage          string             `json:"storage,omitempty"`
	CloudInitStorage string             `json:"cloudinit_storage,omitempty"`
	MemoryMB         int                `json:"memory_mb,omitempty"`
	Sockets          int                `json:"sockets,omitempty"`
	Cores            int                `json:"cores,omitempty"`
	Description      string             `json:"description,omitempty"`
	Net0             string             `json:"net0,omitempty"`
	SCSIHW           string             `json:"scsihw,omitempty"`
	DiskBus          string             `json:"disk_bus,omitempty"`
	BootOrder        string             `json:"boot_order,omitempty"`
	Agent            *bool              `json:"agent,omitempty"`
	SerialConsole    *bool              `json:"serial_console,omitempty"`
	CleanupImage     *bool              `json:"cleanup_image,omitempty"`
}

// swagger:model ProxmoxVMTemplateResult
type ProxmoxVMTemplateResult struct {
	Node             string `json:"node"`
	VMID             int    `json:"vmid"`
	Name             string `json:"name"`
	ImportedVolumeID string `json:"imported_volume_id"`
	Template         bool   `json:"template"`
}

// swagger:model ProxmoxPVEUserCreateRequest
type ProxmoxPVEUserCreateRequest struct {
	SSH          SSHOptions `json:"ssh,omitempty"`
	HostServerID *uuid.UUID `json:"host_server_id,omitempty"`
	Node         string     `json:"node,omitempty"`
	Username     string     `json:"username,omitempty"`
	Realm        string     `json:"realm,omitempty"`
	Comment      string     `json:"comment,omitempty"`
	Password     string     `json:"password,omitempty"`
	Force        bool       `json:"force,omitempty"`
}

// swagger:model ProxmoxPVEUserCreateResult
type ProxmoxPVEUserCreateResult struct {
	Host      string `json:"host"`
	UserID    string `json:"userid"`
	Username  string `json:"username"`
	Realm     string `json:"realm"`
	Created   bool   `json:"created"`
	Recreated bool   `json:"recreated"`
}

// swagger:model ProxmoxAPITokenCreateRequest
type ProxmoxAPITokenCreateRequest struct {
	SSH               SSHOptions `json:"ssh,omitempty"`
	HostServerID      *uuid.UUID `json:"host_server_id,omitempty"`
	Node              string     `json:"node,omitempty"`
	HostURL           string     `json:"host_url,omitempty"`
	UserID            string     `json:"userid,omitempty"`
	Username          string     `json:"username,omitempty"`
	Realm             string     `json:"realm,omitempty"`
	TokenID           string     `json:"token_id,omitempty"`
	Comment           string     `json:"comment,omitempty"`
	Role              string     `json:"role,omitempty"`
	ACLPath           string     `json:"acl_path,omitempty"`
	ExpirationDate    string     `json:"expiration_date,omitempty"`
	DaysValid         int        `json:"days_valid,omitempty"`
	Privsep           bool       `json:"privsep,omitempty"`
	Force             bool       `json:"force,omitempty"`
	Yolo              bool       `json:"yolo,omitempty"`
	Verify            *bool      `json:"verify,omitempty"`
	StoreAsUserSecret *bool      `json:"store_as_user_secret,omitempty"`
}

// swagger:model ProxmoxAPITokenCreateResult
type ProxmoxAPITokenCreateResult struct {
	Host                string     `json:"host"`
	Node                string     `json:"node"`
	HostURL             string     `json:"host_url"`
	UserID              string     `json:"userid"`
	TokenID             string     `json:"token_id"`
	FullTokenID         string     `json:"full_token_id"`
	Secret              string     `json:"secret"`
	APIToken            string     `json:"api_token"`
	Role                string     `json:"role"`
	ACLPath             string     `json:"acl_path"`
	ExpiresAtUnix       int64      `json:"expires_at_unix,omitempty"`
	Privsep             bool       `json:"privsep"`
	Yolo                bool       `json:"yolo"`
	AssignedRoles       []string   `json:"assigned_roles,omitempty"`
	AssignedPrivileges  []string   `json:"assigned_privileges,omitempty"`
	DirectChecks        []string   `json:"direct_checks,omitempty"`
	InferredChecks      []string   `json:"inferred_checks,omitempty"`
	MissingCapabilities []string   `json:"missing_capabilities,omitempty"`
	StoredSecretID      *uuid.UUID `json:"stored_secret_id,omitempty"`
}

// swagger:parameters ListProxmoxVMs
type ListProxmoxVMsParams struct {
	// Host server ID for a Proxmox VE node. When supplied, auth and SSH details are resolved automatically for the current user.
	// in: query
	HostServerID *uuid.UUID `json:"host_server_id,omitempty"`
	// Optional stored Proxmox secret ID to use for this host.
	// in: query
	ProxmoxSecretID *uuid.UUID `json:"proxmox_secret_id,omitempty"`
	// Proxmox node name. Optional when host_server_id resolves the node automatically.
	// in: query
	Node string `json:"node,omitempty"`
	// Whether to include full VM info.
	// in: query
	Full bool `json:"full"`
}

// swagger:parameters ListProxmoxContainers
type ListProxmoxContainersParams struct {
	// Host server ID for a Proxmox VE node. When supplied, auth and SSH details are resolved automatically for the current user.
	// in: query
	HostServerID *uuid.UUID `json:"host_server_id,omitempty"`
	// Optional stored Proxmox secret ID to use for this host.
	// in: query
	ProxmoxSecretID *uuid.UUID `json:"proxmox_secret_id,omitempty"`
	// Proxmox node name. Optional when host_server_id resolves the node automatically.
	// in: query
	Node string `json:"node,omitempty"`
	// Whether to include full container info.
	// in: query
	Full bool `json:"full"`
}

// swagger:parameters ListProxmoxWorkloads
type ListProxmoxWorkloadsParams struct {
	// Host server ID for a Proxmox VE node. When supplied, auth and SSH details are resolved automatically for the current user.
	// in: query
	HostServerID *uuid.UUID `json:"host_server_id,omitempty"`
	// Optional stored Proxmox secret ID to use for this host.
	// in: query
	ProxmoxSecretID *uuid.UUID `json:"proxmox_secret_id,omitempty"`
	// Proxmox node name. Optional when host_server_id resolves the node automatically.
	// in: query
	Node string `json:"node,omitempty"`
	// Whether to include full workload info.
	// in: query
	Full bool `json:"full"`
}

// swagger:parameters StartProxmoxVM
type StartProxmoxVMParams struct {
	// VMID to start.
	// in: path
	// required: true
	VMID int `json:"vmid"`
	// Request body.
	// in: body
	Body ProxmoxVMStartRequest
}

// swagger:parameters CreateProxmoxLXC
type CreateProxmoxLXCParams struct {
	// in: body
	Body ProxmoxLXCRequest
}

// swagger:parameters CreateProxmoxVM
type CreateProxmoxVMParams struct {
	// in: body
	Body ProxmoxVMCreateRequest
}

// swagger:parameters CreateProxmoxVMTemplate
type CreateProxmoxVMTemplateParams struct {
	// in: body
	Body ProxmoxVMTemplateRequest
}

// swagger:parameters CreateProxmoxPVEUser
type CreateProxmoxPVEUserParams struct {
	// in: body
	Body ProxmoxPVEUserCreateRequest
}

// swagger:parameters CreateProxmoxAPIToken
type CreateProxmoxAPITokenParams struct {
	// in: body
	Body ProxmoxAPITokenCreateRequest
}

// swagger:response ProxmoxVMListResponse
type ProxmoxVMListResponse struct {
	// in: body
	Body ProxmoxVMListResult
}

// swagger:response ProxmoxContainerListResponse
type ProxmoxContainerListResponse struct {
	// in: body
	Body ProxmoxContainerListResult
}

// swagger:response ProxmoxWorkloadInventoryResponse
type ProxmoxWorkloadInventoryResponse struct {
	// in: body
	Body ProxmoxWorkloadInventoryResult
}

// swagger:response ProxmoxVMStartResponse
type ProxmoxVMStartResponse struct {
	// in: body
	Body ProxmoxVMStartResult
}

// swagger:response ProxmoxLXCResponse
type ProxmoxLXCResponse struct {
	// in: body
	Body ProxmoxLXCResult
}

// swagger:response ProxmoxVMCreateResponse
type ProxmoxVMCreateResponse struct {
	// in: body
	Body ProxmoxVMCreateResult
}

// swagger:response ProxmoxVMTemplateResponse
type ProxmoxVMTemplateResponse struct {
	// in: body
	Body ProxmoxVMTemplateResult
}

// swagger:response ProxmoxPVEUserCreateResponse
type ProxmoxPVEUserCreateResponse struct {
	// in: body
	Body ProxmoxPVEUserCreateResult
}

// swagger:response ProxmoxAPITokenCreateResponse
type ProxmoxAPITokenCreateResponse struct {
	// in: body
	Body ProxmoxAPITokenCreateResult
}
