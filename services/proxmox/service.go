package proxmox

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	coredeploy "github.com/babbage88/infra-core/deployment"
	coreproxmox "github.com/babbage88/infra-core/proxmox"
)

type Service struct {
	defaultAuth coredeploy.ProxmoxAuthOptions
	defaultNode string
}

func NewServiceFromEnv() *Service {
	useToken := envBool("PROXMOX_USE_TOKEN", true)
	skipTLS := envBool("PROXMOX_SKIP_TLS", true)
	auth := coredeploy.ProxmoxAuthOptions{
		HostURL:    firstNonEmptyEnv("PROXMOX_API_URL", "PROXMOX_HOST", "PROXMOX_HOST_URL"),
		APIToken:   os.Getenv("PROXMOX_API_TOKEN"),
		APITokenID: os.Getenv("PROXMOX_API_TOKEN_ID"),
		APISecret:  os.Getenv("PROXMOX_API_SECRET"),
		Username:   os.Getenv("PROXMOX_USERNAME"),
		Password:   os.Getenv("PROXMOX_PASSWORD"),
		UseToken:   &useToken,
		SkipTLS:    &skipTLS,
	}

	return &Service{
		defaultAuth: auth,
		defaultNode: firstNonEmptyEnv("PROXMOX_NODE", "PVE_NODE"),
	}
}

func (s *Service) ListVMs(ctx context.Context, req coredeploy.ProxmoxVMListRequest) (coredeploy.ProxmoxVMListResult, error) {
	req = s.mergeListDefaults(req)
	if strings.TrimSpace(req.Node) == "" {
		return coredeploy.ProxmoxVMListResult{}, fmt.Errorf("node is required")
	}

	client, err := newCoreClient(req.Auth)
	if err != nil {
		return coredeploy.ProxmoxVMListResult{}, err
	}

	full := true
	if req.Full != nil {
		full = *req.Full
	}
	vms, err := client.ListVMs(ctx, req.Node, full)
	if err != nil {
		return coredeploy.ProxmoxVMListResult{}, fmt.Errorf("list proxmox VMs: %w", err)
	}
	result := make([]coredeploy.ProxmoxVM, 0, len(vms))
	for _, vm := range vms {
		result = append(result, coredeploy.ProxmoxVM{
			VMID:           vm.Vmid,
			Name:           vm.Name,
			Status:         vm.Status,
			CPU:            vm.CPU,
			MaxMem:         vm.MaxMem,
			Mem:            vm.Mem,
			MaxDisk:        vm.MaxDisk,
			Disk:           vm.Disk,
			Node:           vm.Node,
			Tags:           vm.Tags,
			Template:       vm.Template,
			Uptime:         vm.Uptime,
			RunningMachine: vm.RunningMachine,
			RunningQemu:    vm.RunningQemu,
		})
	}
	return coredeploy.ProxmoxVMListResult{Node: req.Node, VMs: result}, nil
}

func (s *Service) ListContainers(ctx context.Context, req coredeploy.ProxmoxVMListRequest) (ProxmoxContainerListResult, error) {
	req = s.mergeListDefaults(req)
	if strings.TrimSpace(req.Node) == "" {
		return ProxmoxContainerListResult{}, fmt.Errorf("node is required")
	}

	client, err := newCoreClient(req.Auth)
	if err != nil {
		return ProxmoxContainerListResult{}, err
	}

	containers, err := client.ListLxcContainers(ctx, req.Node)
	if err != nil {
		return ProxmoxContainerListResult{}, fmt.Errorf("list proxmox containers: %w", err)
	}
	result := make([]ProxmoxContainer, 0, len(containers))
	for _, container := range containers {
		result = append(result, ProxmoxContainer{
			VMID:     container.VmId,
			Name:     firstNonEmpty(container.Name, container.Hostname),
			Status:   container.Status,
			CPU:      container.CPU,
			MaxMem:   container.MaxMem,
			Mem:      container.Mem,
			MaxDisk:  container.MaxDisk,
			Disk:     container.Disk,
			Node:     firstNonEmpty(container.Node, req.Node),
			Tags:     container.Tags,
			Template: container.Template,
			Uptime:   container.Uptime,
		})
	}
	return ProxmoxContainerListResult{Node: req.Node, Containers: result}, nil
}

func (s *Service) ListWorkloads(ctx context.Context, req coredeploy.ProxmoxVMListRequest) (ProxmoxWorkloadInventoryResult, error) {
	vmResult, err := s.ListVMs(ctx, req)
	if err != nil {
		return ProxmoxWorkloadInventoryResult{}, err
	}

	containerResult, err := s.ListContainers(ctx, req)
	if err != nil {
		return ProxmoxWorkloadInventoryResult{}, err
	}

	workloads := make([]ProxmoxWorkload, 0, len(vmResult.VMs)+len(containerResult.Containers))
	for _, vm := range vmResult.VMs {
		workloads = append(workloads, ProxmoxWorkload{
			Kind:     "qemu",
			VMID:     vm.VMID,
			Name:     vm.Name,
			Status:   vm.Status,
			CPU:      vm.CPU,
			MaxMem:   vm.MaxMem,
			Mem:      vm.Mem,
			MaxDisk:  vm.MaxDisk,
			Disk:     vm.Disk,
			Node:     firstNonEmpty(vm.Node, vmResult.Node),
			Tags:     vm.Tags,
			Template: vm.Template,
			Uptime:   vm.Uptime,
		})
	}
	for _, container := range containerResult.Containers {
		workloads = append(workloads, ProxmoxWorkload{
			Kind:     "lxc",
			VMID:     container.VMID,
			Name:     container.Name,
			Status:   container.Status,
			CPU:      container.CPU,
			MaxMem:   container.MaxMem,
			Mem:      container.Mem,
			MaxDisk:  container.MaxDisk,
			Disk:     container.Disk,
			Node:     firstNonEmpty(container.Node, containerResult.Node),
			Tags:     container.Tags,
			Template: container.Template,
			Uptime:   container.Uptime,
		})
	}

	return ProxmoxWorkloadInventoryResult{
		Node:           vmResult.Node,
		VMCount:        len(vmResult.VMs),
		ContainerCount: len(containerResult.Containers),
		WorkloadCount:  len(workloads),
		Workloads:      workloads,
	}, nil
}

func (s *Service) StartVM(ctx context.Context, req coredeploy.ProxmoxVMStartRequest) (coredeploy.ProxmoxVMStartResult, error) {
	req = s.mergeStartDefaults(req)
	if strings.TrimSpace(req.Node) == "" {
		return coredeploy.ProxmoxVMStartResult{}, fmt.Errorf("node is required")
	}
	if req.VMID <= 0 {
		return coredeploy.ProxmoxVMStartResult{}, fmt.Errorf("vmid must be greater than zero")
	}

	client, err := newCoreClient(req.Auth)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, err
	}

	resp, err := client.StartVM(ctx, req.Node, req.VMID)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, fmt.Errorf("start proxmox VM: %w", err)
	}
	upid := firstNonEmptyString(
		stringValue(resp["upid"]),
		stringValue(resp["UPID"]),
		stringValue(resp["data"]),
	)
	return coredeploy.ProxmoxVMStartResult{Node: req.Node, VMID: req.VMID, UPID: upid}, nil
}

func (s *Service) CreateLXC(ctx context.Context, req coredeploy.ProxmoxLXCRequest) (coredeploy.ProxmoxLXCResult, error) {
	req = s.mergeLXCDefaults(req)
	return coredeploy.CreateProxmoxLXC(req)
}

func (s *Service) CreateVM(ctx context.Context, req coredeploy.ProxmoxVMCreateRequest) (coredeploy.ProxmoxVMCreateResult, error) {
	req = s.mergeVMCreateDefaults(req)
	return coredeploy.CreateProxmoxVM(req)
}

func (s *Service) CreateVMTemplate(ctx context.Context, req coredeploy.ProxmoxVMTemplateRequest) (coredeploy.ProxmoxVMTemplateResult, error) {
	req = s.mergeVMTemplateDefaults(req)
	return coredeploy.CreateProxmoxVMTemplate(req)
}

func (s *Service) mergeListDefaults(req coredeploy.ProxmoxVMListRequest) coredeploy.ProxmoxVMListRequest {
	req.Auth = mergeAuthDefaults(req.Auth, s.defaultAuth)
	if strings.TrimSpace(req.Node) == "" {
		req.Node = s.defaultNode
	}
	if req.Full == nil {
		value := true
		req.Full = &value
	}
	return req
}

func (s *Service) mergeStartDefaults(req coredeploy.ProxmoxVMStartRequest) coredeploy.ProxmoxVMStartRequest {
	req.Auth = mergeAuthDefaults(req.Auth, s.defaultAuth)
	if strings.TrimSpace(req.Node) == "" {
		req.Node = s.defaultNode
	}
	return req
}

func (s *Service) mergeLXCDefaults(req coredeploy.ProxmoxLXCRequest) coredeploy.ProxmoxLXCRequest {
	req.Auth = mergeAuthDefaults(req.Auth, s.defaultAuth)
	if strings.TrimSpace(req.Node) == "" {
		req.Node = s.defaultNode
	}
	return req
}

func (s *Service) mergeVMCreateDefaults(req coredeploy.ProxmoxVMCreateRequest) coredeploy.ProxmoxVMCreateRequest {
	req.Auth = mergeAuthDefaults(req.Auth, s.defaultAuth)
	if strings.TrimSpace(req.Node) == "" {
		req.Node = s.defaultNode
	}
	return req
}

func (s *Service) mergeVMTemplateDefaults(req coredeploy.ProxmoxVMTemplateRequest) coredeploy.ProxmoxVMTemplateRequest {
	req.Auth = mergeAuthDefaults(req.Auth, s.defaultAuth)
	if strings.TrimSpace(req.Node) == "" {
		req.Node = s.defaultNode
	}
	return req
}

func mergeAuthDefaults(req coredeploy.ProxmoxAuthOptions, defaults coredeploy.ProxmoxAuthOptions) coredeploy.ProxmoxAuthOptions {
	if strings.TrimSpace(req.HostURL) == "" {
		req.HostURL = defaults.HostURL
	}
	if strings.TrimSpace(req.APIToken) == "" {
		req.APIToken = defaults.APIToken
	}
	if strings.TrimSpace(req.APITokenID) == "" {
		req.APITokenID = defaults.APITokenID
	}
	if strings.TrimSpace(req.APISecret) == "" {
		req.APISecret = defaults.APISecret
	}
	if strings.TrimSpace(req.Username) == "" {
		req.Username = defaults.Username
	}
	if strings.TrimSpace(req.Password) == "" {
		req.Password = defaults.Password
	}
	if req.UseToken == nil {
		req.UseToken = defaults.UseToken
	}
	if req.SkipTLS == nil {
		req.SkipTLS = defaults.SkipTLS
	}
	return req
}

func envBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstNonEmptyString(values ...string) string {
	return firstNonEmpty(values...)
}

func stringValue(value any) string {
	return strings.TrimSpace(fmt.Sprintf("%v", value))
}

func newCoreClient(auth coredeploy.ProxmoxAuthOptions) (*coreproxmox.Client, error) {
	hostURL := strings.TrimSpace(auth.HostURL)
	if hostURL == "" {
		return nil, fmt.Errorf("auth.host_url is required")
	}
	useToken := true
	if auth.UseToken != nil {
		useToken = *auth.UseToken
	}
	skipTLS := true
	if auth.SkipTLS != nil {
		skipTLS = *auth.SkipTLS
	}
	if useToken {
		if apiToken := strings.TrimSpace(auth.APIToken); apiToken != "" {
			return coreproxmox.NewClientTokenString(hostURL, apiToken, skipTLS)
		}
		tokenID := strings.TrimSpace(auth.APITokenID)
		secret := strings.TrimSpace(auth.APISecret)
		if tokenID == "" {
			return nil, fmt.Errorf("auth.api_token_id is required")
		}
		if secret == "" {
			return nil, fmt.Errorf("auth.api_secret is required")
		}
		return coreproxmox.NewClientToken(hostURL, tokenID, secret, skipTLS)
	}
	username := strings.TrimSpace(auth.Username)
	if username == "" {
		return nil, fmt.Errorf("auth.username is required")
	}
	if strings.TrimSpace(auth.Password) == "" {
		return nil, fmt.Errorf("auth.password is required")
	}
	return coreproxmox.NewClient(hostURL, username, auth.Password, skipTLS, false)
}
