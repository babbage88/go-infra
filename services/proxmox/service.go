package proxmox

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/babbage88/go-infra/api/authapi"
	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/services/host_servers"
	"github.com/babbage88/go-infra/services/ssh_key_provider"
	"github.com/babbage88/go-infra/services/user_secrets"
	coredeploy "github.com/babbage88/infra-core/deployment"
	coreproxmox "github.com/babbage88/infra-core/proxmox"
	"github.com/google/uuid"
)

const (
	defaultProxmoxPort = "8006"
	proxmoxAppName     = "proxmox"
)

type Service struct {
	db                 *infra_db_pg.Queries
	hostServerProvider host_servers.HostServerProvider
	sshKeyProvider     ssh_key_provider.SshKeySecretProvider
	secretProvider     user_secrets.UserSecretProvider
	accessResolver     UserAccessResolver
	defaultAuth        coredeploy.ProxmoxAuthOptions
	defaultNode        string
}

func NewService(db *infra_db_pg.Queries, hostServerProvider host_servers.HostServerProvider, sshKeyProvider ssh_key_provider.SshKeySecretProvider, secretProvider user_secrets.UserSecretProvider) *Service {
	useToken := true
	skipTLS := true
	defaultAuth := coredeploy.ProxmoxAuthOptions{
		UseToken: &useToken,
		SkipTLS:  &skipTLS,
	}
	return &Service{
		db:                 db,
		hostServerProvider: hostServerProvider,
		sshKeyProvider:     sshKeyProvider,
		secretProvider:     secretProvider,
		accessResolver:     NewUserAccessResolver(db, hostServerProvider, sshKeyProvider, secretProvider, defaultAuth, ""),
		defaultAuth:        defaultAuth,
	}
}

// NewServiceFromEnv is retained for compatibility, but host-scoped resolution should use NewService.
func NewServiceFromEnv() *Service {
	useToken := true
	skipTLS := true
	return &Service{
		defaultAuth: coredeploy.ProxmoxAuthOptions{
			UseToken: &useToken,
			SkipTLS:  &skipTLS,
		},
	}
}

func (s *Service) ListVMs(ctx context.Context, req coredeploy.ProxmoxVMListRequest) (coredeploy.ProxmoxVMListResult, error) {
	var err error
	req.Auth, _, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return coredeploy.ProxmoxVMListResult{}, err
	}
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
	var err error
	req.Auth, _, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return ProxmoxContainerListResult{}, err
	}
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
	var err error
	req.Auth, _, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, err
	}
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

	upid, err := client.StartVM(ctx, req.Node, req.VMID)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, fmt.Errorf("start proxmox VM: %w", err)
	}
	return coredeploy.ProxmoxVMStartResult{Node: req.Node, VMID: req.VMID, UPID: upid}, nil
}

func (s *Service) StopVM(ctx context.Context, req coredeploy.ProxmoxVMStartRequest) (coredeploy.ProxmoxVMStartResult, error) {
	var err error
	req.Auth, _, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, err
	}
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

	upid, err := client.StopVM(ctx, req.Node, req.VMID)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, fmt.Errorf("stop proxmox VM: %w", err)
	}
	return coredeploy.ProxmoxVMStartResult{Node: req.Node, VMID: req.VMID, UPID: upid}, nil
}

func (s *Service) StartContainer(ctx context.Context, req coredeploy.ProxmoxVMStartRequest) (coredeploy.ProxmoxVMStartResult, error) {
	var err error
	req.Auth, _, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, err
	}
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

	upid, err := client.StartLXCContainer(ctx, req.Node, req.VMID)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, fmt.Errorf("start proxmox LXC: %w", err)
	}
	return coredeploy.ProxmoxVMStartResult{Node: req.Node, VMID: req.VMID, UPID: upid}, nil
}

func (s *Service) StopContainer(ctx context.Context, req coredeploy.ProxmoxVMStartRequest) (coredeploy.ProxmoxVMStartResult, error) {
	var err error
	req.Auth, _, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, err
	}
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

	upid, err := client.StopLXCContainer(ctx, req.Node, req.VMID)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, fmt.Errorf("stop proxmox LXC: %w", err)
	}
	return coredeploy.ProxmoxVMStartResult{Node: req.Node, VMID: req.VMID, UPID: upid}, nil
}

func (s *Service) DeleteVM(ctx context.Context, req coredeploy.ProxmoxVMStartRequest) (coredeploy.ProxmoxVMStartResult, error) {
	var err error
	req.Auth, _, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, err
	}
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

	upid, err := client.DeleteVM(ctx, req.Node, req.VMID)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, fmt.Errorf("delete proxmox VM: %w", err)
	}
	return coredeploy.ProxmoxVMStartResult{Node: req.Node, VMID: req.VMID, UPID: upid}, nil
}

func (s *Service) DeleteContainer(ctx context.Context, req coredeploy.ProxmoxVMStartRequest) (coredeploy.ProxmoxVMStartResult, error) {
	var err error
	req.Auth, _, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, err
	}
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

	upid, err := client.DeleteLXCContainer(ctx, req.Node, req.VMID)
	if err != nil {
		return coredeploy.ProxmoxVMStartResult{}, fmt.Errorf("delete proxmox LXC: %w", err)
	}
	return coredeploy.ProxmoxVMStartResult{Node: req.Node, VMID: req.VMID, UPID: upid}, nil
}

func (s *Service) GetVMHardware(ctx context.Context, req coredeploy.ProxmoxVMStartRequest) (coredeploy.ProxmoxVMHardwareResult, error) {
	var err error
	req.Auth, _, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return coredeploy.ProxmoxVMHardwareResult{}, err
	}
	if strings.TrimSpace(req.Node) == "" {
		return coredeploy.ProxmoxVMHardwareResult{}, fmt.Errorf("node is required")
	}
	if req.VMID <= 0 {
		return coredeploy.ProxmoxVMHardwareResult{}, fmt.Errorf("vmid must be greater than zero")
	}

	client, err := newCoreClient(req.Auth)
	if err != nil {
		return coredeploy.ProxmoxVMHardwareResult{}, err
	}

	cfg, err := client.GetVMConfig(ctx, req.Node, req.VMID)
	if err != nil {
		return coredeploy.ProxmoxVMHardwareResult{}, fmt.Errorf("get proxmox VM config: %w", err)
	}

	result := coredeploy.ProxmoxVMHardwareResult{
		Node: req.Node,
		VMID: req.VMID,
		Name: cfg.Name,
		Raw:  cfg.Raw,
	}
	result.MemoryMB, _ = parseJSONNumberInt(cfg.MemoryMB)
	result.Sockets, _ = parseJSONNumberInt(cfg.Sockets)
	result.Cores, _ = parseJSONNumberInt(cfg.Cores)

	if diskKey, diskValue := findPrimaryVMDisk(cfg.Raw); diskKey != "" {
		result.DiskInterface = diskKey
		result.DiskSize = parseConfigSize(diskValue)
	}

	if netValue, ok := cfg.Raw["net0"]; ok {
		model, mac, bridge, vlan := parseVMNetString(netValue)
		result.NICModel = model
		result.MACAddress = mac
		result.Bridge = bridge
		result.VLANTag = vlan
	}

	return result, nil
}

func (s *Service) UpdateVMHardware(ctx context.Context, req coredeploy.ProxmoxVMHardwareUpdateRequest) (coredeploy.ProxmoxVMHardwareResult, error) {
	var err error
	req.Auth, _, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return coredeploy.ProxmoxVMHardwareResult{}, err
	}
	if strings.TrimSpace(req.Node) == "" {
		return coredeploy.ProxmoxVMHardwareResult{}, fmt.Errorf("node is required")
	}
	if req.VMID <= 0 {
		return coredeploy.ProxmoxVMHardwareResult{}, fmt.Errorf("vmid must be greater than zero")
	}

	client, err := newCoreClient(req.Auth)
	if err != nil {
		return coredeploy.ProxmoxVMHardwareResult{}, err
	}

	current, err := client.GetVMConfig(ctx, req.Node, req.VMID)
	if err != nil {
		return coredeploy.ProxmoxVMHardwareResult{}, fmt.Errorf("get proxmox VM config: %w", err)
	}

	update := &coreproxmox.ProxmoxQemuVmConfig{Raw: make(map[string]string)}
	if req.MemoryMB != nil {
		update.MemoryMB = jsonNumberFromInt(*req.MemoryMB)
	}
	if req.Sockets != nil {
		update.Sockets = jsonNumberFromInt(*req.Sockets)
	}
	if req.Cores != nil {
		update.Cores = jsonNumberFromInt(*req.Cores)
	}

	if req.Bridge != nil || req.VLANTag != nil {
		if currentNet, ok := current.Raw["net0"]; ok {
			_, _, nextBridge, nextVLAN := parseVMNetString(currentNet)
			updatedNet := rewriteVMNetString(
				currentNet,
				firstNonEmptyString(derefString(req.Bridge), nextBridge),
				coalesceVLAN(derefString(req.VLANTag), nextVLAN),
			)
			update.Raw["net0"] = updatedNet
		}
	}

	if len(update.Raw) > 0 || update.MemoryMB != "" || update.Sockets != "" || update.Cores != "" {
		if err := client.UpdateVMConfig(ctx, req.Node, req.VMID, update); err != nil {
			return coredeploy.ProxmoxVMHardwareResult{}, fmt.Errorf("update proxmox VM config: %w", err)
		}
	}

	if req.DiskSizeGB != nil {
		diskKey, diskValue := findPrimaryVMDisk(current.Raw)
		if diskKey == "" {
			return coredeploy.ProxmoxVMHardwareResult{}, fmt.Errorf("unable to determine primary VM disk")
		}
		currentSizeGB, err := parseSizeGiB(parseConfigSize(diskValue))
		if err != nil {
			return coredeploy.ProxmoxVMHardwareResult{}, err
		}
		if *req.DiskSizeGB < currentSizeGB {
			return coredeploy.ProxmoxVMHardwareResult{}, fmt.Errorf("disk size can only be expanded")
		}
		if *req.DiskSizeGB > currentSizeGB {
			delta := *req.DiskSizeGB - currentSizeGB
			if _, err := client.ResizeVMDisk(ctx, req.Node, req.VMID, diskKey, fmt.Sprintf("+%dG", delta)); err != nil {
				return coredeploy.ProxmoxVMHardwareResult{}, fmt.Errorf("resize proxmox VM disk: %w", err)
			}
		}
	}

	return s.GetVMHardware(ctx, coredeploy.ProxmoxVMStartRequest{
		Auth:            req.Auth,
		HostServerID:    req.HostServerID,
		ProxmoxSecretID: req.ProxmoxSecretID,
		Node:            req.Node,
		VMID:            req.VMID,
	})
}

func (s *Service) GetLXCResources(ctx context.Context, req coredeploy.ProxmoxVMStartRequest) (coredeploy.ProxmoxLXCResourcesResult, error) {
	var err error
	req.Auth, _, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return coredeploy.ProxmoxLXCResourcesResult{}, err
	}
	if strings.TrimSpace(req.Node) == "" {
		return coredeploy.ProxmoxLXCResourcesResult{}, fmt.Errorf("node is required")
	}
	if req.VMID <= 0 {
		return coredeploy.ProxmoxLXCResourcesResult{}, fmt.Errorf("vmid must be greater than zero")
	}

	client, err := newCoreClient(req.Auth)
	if err != nil {
		return coredeploy.ProxmoxLXCResourcesResult{}, err
	}

	raw, err := client.GetLXCConfig(ctx, req.Node, req.VMID)
	if err != nil {
		return coredeploy.ProxmoxLXCResourcesResult{}, fmt.Errorf("get proxmox LXC config: %w", err)
	}
	cfg := coreproxmox.ParseLXCConfig(raw, req.VMID)

	result := coredeploy.ProxmoxLXCResourcesResult{
		Node:       req.Node,
		VMID:       req.VMID,
		Hostname:   cfg.Hostname,
		MemoryMB:   cfg.Memory,
		SwapMB:     cfg.Swap,
		Cores:      cfg.Cores,
		RootFS:     raw["rootfs"],
		RootFSSize: parseRootFSSize(raw["rootfs"]),
		Storage:    cfg.Storage,
		Raw:        raw,
	}
	if netValue, ok := raw["net0"]; ok {
		result.Bridge, result.VLANTag = parseLXCNetString(netValue)
	}
	return result, nil
}

func (s *Service) UpdateLXCResources(ctx context.Context, req coredeploy.ProxmoxLXCResourcesUpdateRequest) (coredeploy.ProxmoxLXCResourcesResult, error) {
	var err error
	req.Auth, _, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return coredeploy.ProxmoxLXCResourcesResult{}, err
	}
	if strings.TrimSpace(req.Node) == "" {
		return coredeploy.ProxmoxLXCResourcesResult{}, fmt.Errorf("node is required")
	}
	if req.VMID <= 0 {
		return coredeploy.ProxmoxLXCResourcesResult{}, fmt.Errorf("vmid must be greater than zero")
	}

	client, err := newCoreClient(req.Auth)
	if err != nil {
		return coredeploy.ProxmoxLXCResourcesResult{}, err
	}

	current, err := client.GetLXCConfig(ctx, req.Node, req.VMID)
	if err != nil {
		return coredeploy.ProxmoxLXCResourcesResult{}, fmt.Errorf("get proxmox LXC config: %w", err)
	}
	params := map[string]string{}
	if req.MemoryMB != nil {
		params["memory"] = strconv.Itoa(*req.MemoryMB)
	}
	if req.SwapMB != nil {
		params["swap"] = strconv.Itoa(*req.SwapMB)
	}
	if req.Cores != nil {
		params["cores"] = strconv.Itoa(*req.Cores)
	}
	if req.Bridge != nil || req.VLANTag != nil {
		bridge, vlan := parseLXCNetString(current["net0"])
		params["net0"] = rewriteLXCNetString(
			current["net0"],
			firstNonEmptyString(derefString(req.Bridge), bridge),
			coalesceVLAN(derefString(req.VLANTag), vlan),
		)
	}
	if req.RootFSSizeGB != nil {
		storage, size := parseLXCStorageAndSize(current["rootfs"])
		currentSizeGB, err := parseSizeGiB(size)
		if err != nil {
			return coredeploy.ProxmoxLXCResourcesResult{}, err
		}
		if *req.RootFSSizeGB < currentSizeGB {
			return coredeploy.ProxmoxLXCResourcesResult{}, fmt.Errorf("rootfs size can only be expanded")
		}
		params["rootfs"] = fmt.Sprintf("%s:%dG", storage, *req.RootFSSizeGB)
	}

	if err := client.UpdateLXCConfig(ctx, req.Node, req.VMID, params); err != nil {
		return coredeploy.ProxmoxLXCResourcesResult{}, fmt.Errorf("update proxmox LXC config: %w", err)
	}

	return s.GetLXCResources(ctx, coredeploy.ProxmoxVMStartRequest{
		Auth:            req.Auth,
		HostServerID:    req.HostServerID,
		ProxmoxSecretID: req.ProxmoxSecretID,
		Node:            req.Node,
		VMID:            req.VMID,
	})
}

func (s *Service) CreateLXC(ctx context.Context, req coredeploy.ProxmoxLXCRequest) (coredeploy.ProxmoxLXCResult, error) {
	var err error
	req.Auth, _, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, coredeploy.SSHOptions{}, req.Node)
	if err != nil {
		return coredeploy.ProxmoxLXCResult{}, err
	}
	return coredeploy.CreateProxmoxLXC(req)
}

func (s *Service) CreateVM(ctx context.Context, req coredeploy.ProxmoxVMCreateRequest) (coredeploy.ProxmoxVMCreateResult, error) {
	var err error
	req.Auth, req.SSH, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, req.SSH, req.Node)
	if err != nil {
		return coredeploy.ProxmoxVMCreateResult{}, err
	}
	return coredeploy.CreateProxmoxVM(req)
}

func (s *Service) CreateVMTemplate(ctx context.Context, req coredeploy.ProxmoxVMTemplateRequest) (coredeploy.ProxmoxVMTemplateResult, error) {
	var err error
	req.Auth, req.SSH, req.Node, err = s.resolveAccess(ctx, req.HostServerID, req.ProxmoxSecretID, req.Auth, req.SSH, req.Node)
	if err != nil {
		return coredeploy.ProxmoxVMTemplateResult{}, err
	}
	return coredeploy.CreateProxmoxVMTemplate(req)
}

func (s *Service) CreatePVEUser(ctx context.Context, req coreproxmox.CreatePVEUserRequest) (coreproxmox.CreatePVEUserResult, error) {
	var err error
	_, req.SSH, req.Node, err = s.resolveSSHAccess(ctx, req.HostServerID, coredeploy.ProxmoxAuthOptions{}, req.SSH, req.Node)
	if err != nil {
		return coreproxmox.CreatePVEUserResult{}, err
	}
	if strings.TrimSpace(req.Node) == "" {
		req.Node = req.SSH.Host
	}
	return coreproxmox.CreatePVEUser(req)
}

func (s *Service) CreateAPIToken(ctx context.Context, req coreproxmox.CreateAPITokenRequest) (coreproxmox.CreateAPITokenResult, error) {
	var err error
	auth, ssh, node, err := s.resolveSSHAccess(ctx, req.HostServerID, coredeploy.ProxmoxAuthOptions{HostURL: req.HostURL}, req.SSH, req.Node)
	if err != nil {
		return coreproxmox.CreateAPITokenResult{}, err
	}
	req.SSH = ssh
	req.Node = node
	req.HostURL = auth.HostURL

	result, err := coreproxmox.CreateAPIToken(req)
	if err != nil {
		return coreproxmox.CreateAPITokenResult{}, err
	}

	shouldStore := req.HostServerID != nil
	if req.StoreAsUserSecret != nil {
		shouldStore = *req.StoreAsUserSecret
	}
	if !shouldStore {
		return result, nil
	}

	userID, err := authapi.GetUserIDFromContext(ctx)
	if err != nil {
		return coreproxmox.CreateAPITokenResult{}, err
	}

	appID, err := s.ensureExternalAppID(ctx, proxmoxAppName)
	if err != nil {
		return coreproxmox.CreateAPITokenResult{}, fmt.Errorf("ensure proxmox external application: %w", err)
	}

	var expiry time.Time
	if result.ExpiresAtUnix > 0 {
		expiry = time.Unix(result.ExpiresAtUnix, 0).UTC()
	}

	secretID, err := s.secretProvider.StoreSecretWithMetadata(
		result.APIToken,
		userID,
		appID,
		expiry,
		user_secrets.SecretMetadata{HostServerID: req.HostServerID},
	)
	if err != nil {
		return coreproxmox.CreateAPITokenResult{}, fmt.Errorf("store proxmox API token for host: %w", err)
	}
	result.StoredSecretID = &secretID
	return result, nil
}

func jsonNumberFromInt(value int) json.Number {
	return json.Number(strconv.Itoa(value))
}

func parseJSONNumberInt(value json.Number) (int, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value.String())
}

func parseVMNetString(value string) (model string, mac string, bridge string, vlan string) {
	parts := strings.Split(value, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if key, val, ok := strings.Cut(part, "="); ok {
			switch key {
			case "bridge":
				bridge = val
			case "tag":
				vlan = val
			default:
				if model == "" && !strings.Contains(key, "rate") {
					model = key
					mac = val
				}
			}
		}
	}
	return
}

func rewriteVMNetString(value string, bridge string, vlan string) string {
	parts := strings.Split(value, ",")
	foundBridge := false
	foundTag := false
	for i, part := range parts {
		key, _, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "bridge":
			parts[i] = "bridge=" + bridge
			foundBridge = true
		case "tag":
			foundTag = true
			if strings.TrimSpace(vlan) == "" {
				parts[i] = ""
			} else {
				parts[i] = "tag=" + vlan
			}
		}
	}
	filtered := make([]string, 0, len(parts)+1)
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			filtered = append(filtered, part)
		}
	}
	if !foundBridge && strings.TrimSpace(bridge) != "" {
		filtered = append(filtered, "bridge="+bridge)
	}
	if !foundTag && strings.TrimSpace(vlan) != "" {
		filtered = append(filtered, "tag="+vlan)
	}
	return strings.Join(filtered, ",")
}

func parseLXCNetString(value string) (bridge string, vlan string) {
	parts := strings.Split(value, ",")
	for _, part := range parts {
		key, val, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "bridge":
			bridge = val
		case "tag":
			vlan = val
		}
	}
	return
}

func rewriteLXCNetString(value string, bridge string, vlan string) string {
	parts := strings.Split(value, ",")
	foundBridge := false
	foundTag := false
	for i, part := range parts {
		key, _, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "bridge":
			parts[i] = "bridge=" + bridge
			foundBridge = true
		case "tag":
			foundTag = true
			if strings.TrimSpace(vlan) == "" {
				parts[i] = ""
			} else {
				parts[i] = "tag=" + vlan
			}
		}
	}
	filtered := make([]string, 0, len(parts)+1)
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			filtered = append(filtered, part)
		}
	}
	if !foundBridge && strings.TrimSpace(bridge) != "" {
		filtered = append(filtered, "bridge="+bridge)
	}
	if !foundTag && strings.TrimSpace(vlan) != "" {
		filtered = append(filtered, "tag="+vlan)
	}
	return strings.Join(filtered, ",")
}

func findPrimaryVMDisk(raw map[string]string) (string, string) {
	prefixes := []string{"scsi", "virtio", "sata", "ide"}
	for _, prefix := range prefixes {
		for key, value := range raw {
			if key == "ide2" || strings.HasPrefix(key, "unused") || strings.HasPrefix(key, "efidisk") || strings.HasPrefix(key, "tpmstate") {
				continue
			}
			if strings.HasPrefix(key, prefix) && strings.Contains(value, "size=") {
				return key, value
			}
		}
	}
	return "", ""
}

func parseConfigSize(value string) string {
	for _, part := range strings.Split(value, ",") {
		key, val, ok := strings.Cut(strings.TrimSpace(part), "=")
		if ok && key == "size" {
			return val
		}
	}
	return ""
}

func parseRootFSSize(value string) string {
	if idx := strings.Index(value, ":"); idx >= 0 && idx < len(value)-1 {
		return value[idx+1:]
	}
	return value
}

func parseLXCStorageAndSize(rootfs string) (string, string) {
	if idx := strings.Index(rootfs, ":"); idx >= 0 && idx < len(rootfs)-1 {
		return rootfs[:idx], rootfs[idx+1:]
	}
	return "", rootfs
}

func parseSizeGiB(value string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("size is missing")
	}
	upper := strings.ToUpper(strings.TrimSpace(value))
	switch {
	case strings.HasSuffix(upper, "G"):
		return strconv.Atoi(strings.TrimSuffix(upper, "G"))
	case strings.HasSuffix(upper, "M"):
		mb, err := strconv.Atoi(strings.TrimSuffix(upper, "M"))
		if err != nil {
			return 0, err
		}
		gb := mb / 1024
		if mb%1024 != 0 {
			gb++
		}
		return gb, nil
	case strings.HasSuffix(upper, "T"):
		tb, err := strconv.Atoi(strings.TrimSuffix(upper, "T"))
		if err != nil {
			return 0, err
		}
		return tb * 1024, nil
	default:
		return strconv.Atoi(upper)
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func coalesceVLAN(requested string, current string) string {
	if requested == "" && current != "" {
		return current
	}
	return requested
}

func (s *Service) resolveAccess(ctx context.Context, hostServerID, proxmoxSecretID *uuid.UUID, auth coredeploy.ProxmoxAuthOptions, ssh coredeploy.SSHOptions, node string) (coredeploy.ProxmoxAuthOptions, coredeploy.SSHOptions, string, error) {
	if s.accessResolver == nil {
		auth = mergeAuthDefaults(auth, s.defaultAuth)
		auth = ensureAuthBooleans(auth)
		if strings.TrimSpace(node) == "" {
			node = s.defaultNode
		}
		return auth, ssh, node, nil
	}

	resolution, err := s.accessResolver.ResolveAccess(ctx, hostServerID, proxmoxSecretID, auth, ssh, node)
	if err != nil {
		return auth, ssh, node, err
	}

	return resolution.Auth, resolution.SSH, resolution.Node, nil
}

func (s *Service) resolveSSHAccess(ctx context.Context, hostServerID *uuid.UUID, auth coredeploy.ProxmoxAuthOptions, ssh coredeploy.SSHOptions, node string) (coredeploy.ProxmoxAuthOptions, coredeploy.SSHOptions, string, error) {
	if s.accessResolver == nil {
		auth = mergeAuthDefaults(auth, s.defaultAuth)
		auth = ensureAuthBooleans(auth)
		if strings.TrimSpace(node) == "" {
			node = s.defaultNode
		}
		return auth, ssh, node, nil
	}

	resolution, err := s.accessResolver.ResolveSSHAccess(ctx, hostServerID, auth, ssh, node)
	if err != nil {
		return auth, ssh, node, err
	}

	return resolution.Auth, resolution.SSH, resolution.Node, nil
}

func (s *Service) resolveConsoleCookieAccess(ctx context.Context, hostServerID *uuid.UUID, auth coredeploy.ProxmoxAuthOptions, ssh coredeploy.SSHOptions, node string) (coredeploy.ProxmoxAuthOptions, coredeploy.SSHOptions, string, error) {
	if s.accessResolver == nil {
		auth = mergeAuthDefaults(auth, s.defaultAuth)
		auth = ensureAuthBooleans(auth)
		useToken := false
		auth.UseToken = &useToken
		auth.APIToken = ""
		auth.APITokenID = ""
		auth.APISecret = ""
		auth.Username = normalizeProxmoxPasswordUser(auth.Username)
		if strings.TrimSpace(auth.Username) == "" {
			return coredeploy.ProxmoxAuthOptions{}, coredeploy.SSHOptions{}, "", fmt.Errorf("auth.username is required for cookie-based Proxmox console auth")
		}
		if strings.TrimSpace(auth.Password) == "" {
			return coredeploy.ProxmoxAuthOptions{}, coredeploy.SSHOptions{}, "", fmt.Errorf("auth.password is required for cookie-based Proxmox console auth")
		}
		if strings.TrimSpace(node) == "" {
			node = s.defaultNode
		}
		return auth, ssh, node, nil
	}

	resolution, err := s.accessResolver.ResolveConsoleCookieAccess(ctx, hostServerID, auth, ssh, node)
	if err != nil {
		return auth, ssh, node, err
	}

	return resolution.Auth, resolution.SSH, resolution.Node, nil
}

func normalizeProxmoxPasswordUser(username string) string {
	username = strings.TrimSpace(username)
	if username == "" || strings.Contains(username, "@") {
		return username
	}
	return username + "@pam"
}

func (s *Service) ensureExternalAppID(ctx context.Context, name string) (uuid.UUID, error) {
	if s.accessResolver != nil {
		return s.accessResolver.EnsureExternalAppID(ctx, name)
	}
	return s.db.GetExternalAppIdByName(ctx, name)
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

func ensureAuthBooleans(auth coredeploy.ProxmoxAuthOptions) coredeploy.ProxmoxAuthOptions {
	if auth.UseToken == nil {
		useToken := true
		auth.UseToken = &useToken
	}
	if auth.SkipTLS == nil {
		skipTLS := true
		auth.SkipTLS = &skipTLS
	}
	return auth
}

func hasUsableProxmoxAuth(auth coredeploy.ProxmoxAuthOptions) bool {
	if strings.TrimSpace(auth.APIToken) != "" {
		return true
	}
	if strings.TrimSpace(auth.APITokenID) != "" && strings.TrimSpace(auth.APISecret) != "" {
		return true
	}
	return strings.TrimSpace(auth.Username) != "" && strings.TrimSpace(auth.Password) != ""
}

func hasExplicitSSHKeySource(ssh coredeploy.SSHOptions) bool {
	return strings.TrimSpace(ssh.KeyPath) != "" || strings.TrimSpace(ssh.PrivateKeyPEM) != "" || strings.TrimSpace(ssh.PrivateKeyBase64) != ""
}

func isProxmoxPlatformHost(host *host_servers.HostServer) bool {
	for _, platform := range host.PlatformTypes {
		if strings.Contains(strings.ToLower(platform.Name), "proxmox") {
			return true
		}
	}
	return false
}

func hostAddress(host *host_servers.HostServer) string {
	if strings.TrimSpace(host.Hostname) != "" {
		return strings.TrimSpace(host.Hostname)
	}
	if host.IPAddress != nil {
		return host.IPAddress.String()
	}
	return ""
}

func buildProxmoxHostURL(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return host
	}
	return "https://" + host + ":" + defaultProxmoxPort
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
