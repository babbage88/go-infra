package proxmox

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

// swagger:model ProxmoxVMStartRequest
type ProxmoxVMStartRequest struct {
	Auth ProxmoxAuthOptions `json:"auth,omitempty"`
	Node string             `json:"node,omitempty"`
	VMID int                `json:"vmid,omitempty"`
}

// swagger:model ProxmoxVMStartResult
type ProxmoxVMStartResult struct {
	Node string `json:"node"`
	VMID int    `json:"vmid"`
	UPID string `json:"upid,omitempty"`
}

// swagger:parameters ListProxmoxVMs
type ListProxmoxVMsParams struct {
	// Proxmox node name.
	// in: query
	Node string `json:"node"`
	// Whether to include full VM info.
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

// swagger:response ProxmoxVMListResponse
type ProxmoxVMListResponse struct {
	// in: body
	Body ProxmoxVMListResult
}

// swagger:response ProxmoxVMStartResponse
type ProxmoxVMStartResponse struct {
	// in: body
	Body ProxmoxVMStartResult
}
