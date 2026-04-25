package proxmox

import (
	"context"
	"fmt"
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
	_, req.SSH, req.Node, err = s.resolveAccess(ctx, req.HostServerID, nil, coredeploy.ProxmoxAuthOptions{}, req.SSH, req.Node)
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
	auth, ssh, node, err := s.resolveAccess(ctx, req.HostServerID, nil, coredeploy.ProxmoxAuthOptions{HostURL: req.HostURL}, req.SSH, req.Node)
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

func (s *Service) resolveAccess(ctx context.Context, hostServerID, proxmoxSecretID *uuid.UUID, auth coredeploy.ProxmoxAuthOptions, ssh coredeploy.SSHOptions, node string) (coredeploy.ProxmoxAuthOptions, coredeploy.SSHOptions, string, error) {
	auth = mergeAuthDefaults(auth, s.defaultAuth)
	auth = ensureAuthBooleans(auth)
	if strings.TrimSpace(node) == "" {
		node = s.defaultNode
	}
	if hostServerID == nil {
		return auth, ssh, node, nil
	}
	if s.db == nil || s.hostServerProvider == nil || s.secretProvider == nil {
		return auth, ssh, node, fmt.Errorf("host-scoped proxmox resolution is unavailable")
	}

	userID, err := authapi.GetUserIDFromContext(ctx)
	if err != nil {
		return auth, ssh, node, err
	}

	hostServer, err := s.hostServerProvider.GetHostServer(ctx, *hostServerID)
	if err != nil {
		return auth, ssh, node, fmt.Errorf("resolve host server %s: %w", hostServerID.String(), err)
	}
	if !isProxmoxPlatformHost(hostServer) {
		return auth, ssh, node, fmt.Errorf("host server %s is not marked as a Proxmox platform", hostServer.ID)
	}

	hostAddr := hostAddress(hostServer)
	if hostAddr == "" {
		return auth, ssh, node, fmt.Errorf("host server %s is missing a hostname or IP address", hostServer.ID)
	}
	if strings.TrimSpace(auth.HostURL) == "" {
		auth.HostURL = buildProxmoxHostURL(hostAddr)
	}
	if strings.TrimSpace(node) == "" {
		node = strings.TrimSpace(hostServer.Hostname)
	}

	mapping, err := s.getUserSSHMappingForHost(ctx, *hostServerID, userID)
	if err != nil {
		return auth, ssh, node, err
	}
	ssh, err = s.populateSSHFromMapping(ctx, ssh, hostAddr, mapping)
	if err != nil {
		return auth, ssh, node, err
	}

	if hasUsableProxmoxAuth(auth) {
		return auth, ssh, node, nil
	}

	secret, err := s.resolveProxmoxTokenSecret(ctx, userID, *hostServerID, proxmoxSecretID)
	if err != nil {
		return auth, ssh, node, err
	}
	auth.APIToken = string(secret.ExternalAuthToken.Token)
	return auth, ssh, node, nil
}

func (s *Service) getUserSSHMappingForHost(ctx context.Context, hostServerID, userID uuid.UUID) (*infra_db_pg.UserSshKeyMapping, error) {
	mappings, err := s.db.GetSSHKeyHostMappingsByHostId(ctx, hostServerID)
	if err != nil {
		return nil, fmt.Errorf("get SSH key mappings for host %s: %w", hostServerID, err)
	}
	for _, mapping := range mappings {
		if mapping.UserID == userID {
			return &mapping, nil
		}
	}
	return nil, fmt.Errorf("no SSH key mapping found for user %s on host %s", userID, hostServerID)
}

func (s *Service) populateSSHFromMapping(ctx context.Context, ssh coredeploy.SSHOptions, hostAddr string, mapping *infra_db_pg.UserSshKeyMapping) (coredeploy.SSHOptions, error) {
	if strings.TrimSpace(ssh.Host) == "" {
		ssh.Host = hostAddr
	}
	if strings.TrimSpace(ssh.User) == "" {
		ssh.User = strings.TrimSpace(mapping.HostserverUsername)
	}
	if ssh.Port == 0 {
		ssh.Port = 22
	}
	if hasExplicitSSHKeySource(ssh) {
		return ssh, nil
	}

	key, err := s.db.GetSSHKeyById(ctx, mapping.SshKeyID)
	if err != nil {
		return ssh, fmt.Errorf("get SSH key %s for proxmox host access: %w", mapping.SshKeyID, err)
	}

	privateKey, err := s.secretProvider.RetrieveSecret(key.PrivSecretID)
	if err != nil {
		return ssh, fmt.Errorf("retrieve SSH private key secret %s: %w", key.PrivSecretID, err)
	}
	ssh.PrivateKeyPEM = string(privateKey.ExternalAuthToken.Token)

	if key.PassphraseID != nil && strings.TrimSpace(ssh.Passphrase) == "" {
		passphrase, err := s.secretProvider.RetrieveSecret(*key.PassphraseID)
		if err != nil {
			return ssh, fmt.Errorf("retrieve SSH passphrase secret %s: %w", key.PassphraseID.String(), err)
		}
		ssh.Passphrase = string(passphrase.ExternalAuthToken.Token)
	}

	return ssh, nil
}

func (s *Service) resolveProxmoxTokenSecret(ctx context.Context, userID, hostServerID uuid.UUID, secretID *uuid.UUID) (*user_secrets.RetrievedUserSecret, error) {
	if secretID != nil {
		secret, err := s.secretProvider.RetrieveSecret(*secretID)
		if err != nil {
			return nil, fmt.Errorf("retrieve proxmox secret %s: %w", secretID.String(), err)
		}
		if secret.ExternalAuthToken.UserID != userID {
			return nil, fmt.Errorf("proxmox secret %s does not belong to the current user", secretID.String())
		}
		if secret.Metadata.HostServerID != nil && *secret.Metadata.HostServerID != hostServerID {
			return nil, fmt.Errorf("proxmox secret %s is not mapped to host %s", secretID.String(), hostServerID)
		}
		return secret, nil
	}

	appID, err := s.ensureExternalAppID(ctx, proxmoxAppName)
	if err != nil {
		return nil, err
	}
	secrets, err := s.db.GetExternalAuthTokensByUserIdAndAppId(ctx, infra_db_pg.GetExternalAuthTokensByUserIdAndAppIdParams{
		UserID:        userID,
		ExternalAppID: appID,
	})
	if err != nil {
		return nil, fmt.Errorf("list proxmox secrets for user %s: %w", userID, err)
	}
	sort.SliceStable(secrets, func(i, j int) bool {
		return secrets[i].CreatedAt.Time.After(secrets[j].CreatedAt.Time)
	})

	for _, token := range secrets {
		secret, err := s.secretProvider.RetrieveSecret(token.ID)
		if err != nil {
			continue
		}
		if secret.Metadata.HostServerID != nil && *secret.Metadata.HostServerID == hostServerID {
			return secret, nil
		}
	}

	return nil, fmt.Errorf("no proxmox API token stored for host %s; create one with infractl proxmox new api-token and map it to this host", hostServerID)
}

func (s *Service) ensureExternalAppID(ctx context.Context, name string) (uuid.UUID, error) {
	appID, err := s.db.GetExternalAppIdByName(ctx, name)
	if err == nil {
		return appID, nil
	}
	if err != pgx.ErrNoRows {
		return uuid.Nil, err
	}

	created, createErr := s.db.InsertExternalAppIntegrationByName(ctx, infra_db_pg.InsertExternalAppIntegrationByNameParams{
		ID:   uuid.New(),
		Name: name,
	})
	if createErr == nil {
		return created.ID, nil
	}

	appID, err = s.db.GetExternalAppIdByName(ctx, name)
	if err != nil {
		return uuid.Nil, err
	}
	return appID, nil
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
