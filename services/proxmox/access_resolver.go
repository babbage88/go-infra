package proxmox

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/babbage88/go-infra/api/authapi"
	"github.com/babbage88/go-infra/database/infra_db_pg"
	"github.com/babbage88/go-infra/services/host_servers"
	"github.com/babbage88/go-infra/services/ssh_key_provider"
	"github.com/babbage88/go-infra/services/user_secrets"
	coredeploy "github.com/babbage88/infra-core/deployment"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AccessResolution struct {
	Auth coredeploy.ProxmoxAuthOptions
	SSH  coredeploy.SSHOptions
	Node string
}

type UserAccessResolver interface {
	ResolveAccess(ctx context.Context, hostServerID, proxmoxSecretID *uuid.UUID, auth coredeploy.ProxmoxAuthOptions, ssh coredeploy.SSHOptions, node string) (AccessResolution, error)
	ResolveSSHAccess(ctx context.Context, hostServerID *uuid.UUID, auth coredeploy.ProxmoxAuthOptions, ssh coredeploy.SSHOptions, node string) (AccessResolution, error)
	EnsureExternalAppID(ctx context.Context, name string) (uuid.UUID, error)
}

type UserAccessResolverImpl struct {
	db                 *infra_db_pg.Queries
	hostServerProvider host_servers.HostServerProvider
	sshKeyProvider     ssh_key_provider.SshKeySecretProvider
	secretProvider     user_secrets.UserSecretProvider
	defaultAuth        coredeploy.ProxmoxAuthOptions
	defaultNode        string
}

func NewUserAccessResolver(db *infra_db_pg.Queries, hostServerProvider host_servers.HostServerProvider, sshKeyProvider ssh_key_provider.SshKeySecretProvider, secretProvider user_secrets.UserSecretProvider, defaultAuth coredeploy.ProxmoxAuthOptions, defaultNode string) *UserAccessResolverImpl {
	return &UserAccessResolverImpl{
		db:                 db,
		hostServerProvider: hostServerProvider,
		sshKeyProvider:     sshKeyProvider,
		secretProvider:     secretProvider,
		defaultAuth:        defaultAuth,
		defaultNode:        defaultNode,
	}
}

func (r *UserAccessResolverImpl) ResolveAccess(ctx context.Context, hostServerID, proxmoxSecretID *uuid.UUID, auth coredeploy.ProxmoxAuthOptions, ssh coredeploy.SSHOptions, node string) (AccessResolution, error) {
	return r.resolveHostAccess(ctx, hostServerID, proxmoxSecretID, auth, ssh, node, true)
}

func (r *UserAccessResolverImpl) ResolveSSHAccess(ctx context.Context, hostServerID *uuid.UUID, auth coredeploy.ProxmoxAuthOptions, ssh coredeploy.SSHOptions, node string) (AccessResolution, error) {
	return r.resolveHostAccess(ctx, hostServerID, nil, auth, ssh, node, false)
}

func (r *UserAccessResolverImpl) resolveHostAccess(ctx context.Context, hostServerID, proxmoxSecretID *uuid.UUID, auth coredeploy.ProxmoxAuthOptions, ssh coredeploy.SSHOptions, node string, requireProxmoxAuth bool) (AccessResolution, error) {
	auth = mergeAuthDefaults(auth, r.defaultAuth)
	auth = ensureAuthBooleans(auth)
	if strings.TrimSpace(node) == "" {
		node = r.defaultNode
	}
	if hostServerID == nil {
		return AccessResolution{Auth: auth, SSH: ssh, Node: node}, nil
	}
	if r.db == nil || r.hostServerProvider == nil || r.secretProvider == nil || r.sshKeyProvider == nil {
		return AccessResolution{}, fmt.Errorf("host-scoped proxmox resolution is unavailable")
	}

	userID, err := authapi.GetUserIDFromContext(ctx)
	if err != nil {
		return AccessResolution{}, err
	}

	hostServer, err := r.hostServerProvider.GetHostServer(ctx, *hostServerID)
	if err != nil {
		return AccessResolution{}, fmt.Errorf("resolve host server %s: %w", hostServerID.String(), err)
	}
	if !isProxmoxPlatformHost(hostServer) {
		return AccessResolution{}, fmt.Errorf("host server %s is not marked as a Proxmox platform", hostServer.ID)
	}

	hostAddr := hostAddress(hostServer)
	if hostAddr == "" {
		return AccessResolution{}, fmt.Errorf("host server %s is missing a hostname or IP address", hostServer.ID)
	}
	if strings.TrimSpace(auth.HostURL) == "" {
		auth.HostURL = buildProxmoxHostURL(hostAddr)
	}
	if strings.TrimSpace(node) == "" {
		node = strings.TrimSpace(hostServer.Hostname)
	}

	mapping, err := r.getUserSSHMappingForHost(userID, *hostServerID)
	if err != nil {
		return AccessResolution{}, err
	}
	ssh, err = r.populateSSHFromMapping(ssh, hostAddr, mapping)
	if err != nil {
		return AccessResolution{}, err
	}

	if !requireProxmoxAuth || hasUsableProxmoxAuth(auth) {
		return AccessResolution{Auth: auth, SSH: ssh, Node: node}, nil
	}

	secret, err := r.resolveProxmoxTokenSecret(ctx, userID, *hostServerID, proxmoxSecretID)
	if err != nil {
		return AccessResolution{}, err
	}
	auth.APIToken = string(secret.ExternalAuthToken.Token)

	return AccessResolution{Auth: auth, SSH: ssh, Node: node}, nil
}

func (r *UserAccessResolverImpl) getUserSSHMappingForHost(userID, hostServerID uuid.UUID) (*ssh_key_provider.CreateSshKeyHostMappingResult, error) {
	mappings, err := r.sshKeyProvider.GetSshKeyHostMappingsByHostId(hostServerID)
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

func (r *UserAccessResolverImpl) populateSSHFromMapping(ssh coredeploy.SSHOptions, hostAddr string, mapping *ssh_key_provider.CreateSshKeyHostMappingResult) (coredeploy.SSHOptions, error) {
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

	key, err := r.sshKeyProvider.GetSshKeyById(mapping.SshKeyID)
	if err != nil {
		return ssh, fmt.Errorf("get SSH key %s for proxmox host access: %w", mapping.SshKeyID, err)
	}

	privateKey, err := r.secretProvider.RetrieveSecret(key.PrivateKeyId)
	if err != nil {
		return ssh, fmt.Errorf("retrieve SSH private key secret %s: %w", key.PrivateKeyId, err)
	}
	ssh.PrivateKeyPEM = string(privateKey.ExternalAuthToken.Token)

	if key.PassphraseSecretId != nil && strings.TrimSpace(ssh.Passphrase) == "" {
		passphrase, err := r.secretProvider.RetrieveSecret(*key.PassphraseSecretId)
		if err != nil {
			return ssh, fmt.Errorf("retrieve SSH passphrase secret %s: %w", key.PassphraseSecretId.String(), err)
		}
		ssh.Passphrase = string(passphrase.ExternalAuthToken.Token)
	}

	return ssh, nil
}

func (r *UserAccessResolverImpl) resolveProxmoxTokenSecret(ctx context.Context, userID, hostServerID uuid.UUID, secretID *uuid.UUID) (*user_secrets.RetrievedUserSecret, error) {
	if secretID != nil {
		secret, err := r.secretProvider.RetrieveSecret(*secretID)
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

	appID, err := r.EnsureExternalAppID(ctx, proxmoxAppName)
	if err != nil {
		return nil, err
	}
	secrets, err := r.db.GetExternalAuthTokensByUserIdAndAppId(ctx, infra_db_pg.GetExternalAuthTokensByUserIdAndAppIdParams{
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
		secret, err := r.secretProvider.RetrieveSecret(token.ID)
		if err != nil {
			continue
		}
		if secret.Metadata.HostServerID != nil && *secret.Metadata.HostServerID == hostServerID {
			return secret, nil
		}
	}

	return nil, fmt.Errorf("no proxmox API token stored for host %s; create one with infractl proxmox new api-token and map it to this host", hostServerID)
}

func (r *UserAccessResolverImpl) EnsureExternalAppID(ctx context.Context, name string) (uuid.UUID, error) {
	appID, err := r.db.GetExternalAppIdByName(ctx, name)
	if err == nil {
		return appID, nil
	}
	if err != pgx.ErrNoRows {
		return uuid.Nil, err
	}

	created, createErr := r.db.InsertExternalAppIntegrationByName(ctx, infra_db_pg.InsertExternalAppIntegrationByNameParams{
		ID:   uuid.New(),
		Name: name,
	})
	if createErr == nil {
		return created.ID, nil
	}

	appID, err = r.db.GetExternalAppIdByName(ctx, name)
	if err != nil {
		return uuid.Nil, err
	}
	return appID, nil
}
