package proxmox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	coredeploy "github.com/babbage88/infra-core/deployment"
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

	client, err := newClient(req.Auth)
	if err != nil {
		return coredeploy.ProxmoxVMListResult{}, err
	}

	full := true
	if req.Full != nil {
		full = *req.Full
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var vms []coredeploy.ProxmoxVM
	path := fmt.Sprintf("%s/%s/qemu?full=%d", apiNodesPath, url.PathEscape(req.Node), boolInt(full))
	if err := client.do(ctx, http.MethodGet, path, nil, &vms); err != nil {
		return coredeploy.ProxmoxVMListResult{}, fmt.Errorf("list proxmox VMs: %w", err)
	}

	return coredeploy.ProxmoxVMListResult{Node: req.Node, VMs: vms}, nil
}

func (s *Service) ListContainers(ctx context.Context, req coredeploy.ProxmoxVMListRequest) (ProxmoxContainerListResult, error) {
	req = s.mergeListDefaults(req)
	if strings.TrimSpace(req.Node) == "" {
		return ProxmoxContainerListResult{}, fmt.Errorf("node is required")
	}

	client, err := newClient(req.Auth)
	if err != nil {
		return ProxmoxContainerListResult{}, err
	}

	full := true
	if req.Full != nil {
		full = *req.Full
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var containers []ProxmoxContainer
	path := fmt.Sprintf("%s/%s/lxc?full=%d", apiNodesPath, url.PathEscape(req.Node), boolInt(full))
	if err := client.do(ctx, http.MethodGet, path, nil, &containers); err != nil {
		return ProxmoxContainerListResult{}, fmt.Errorf("list proxmox containers: %w", err)
	}

	for i := range containers {
		containers[i].Node = req.Node
	}

	return ProxmoxContainerListResult{Node: req.Node, Containers: containers}, nil
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

	client, err := newClient(req.Auth)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var upid string
	path := fmt.Sprintf("%s/%s/qemu/%d/status/start", apiNodesPath, url.PathEscape(req.Node), req.VMID)
	if err := client.do(ctx, http.MethodPost, path, nil, &upid); err != nil {
		return coredeploy.ProxmoxVMStartResult{}, fmt.Errorf("start proxmox VM: %w", err)
	}

	return coredeploy.ProxmoxVMStartResult{Node: req.Node, VMID: req.VMID, UPID: upid}, nil
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

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
