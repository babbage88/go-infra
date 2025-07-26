-- name: CreateUser :one
INSERT INTO users (
    username,
    password,
    email
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetUserById :one
SELECT
    "id",
    "username",
    "password",
    "email",
    "roles",
    "role_ids",
    "created_at",
    "last_modified",
    "enabled",
    "is_deleted"
FROM public.users_with_roles uwr
WHERE "id" = $1;

-- name: GetUserByName :one
SELECT
    "id",
    "username",
    "password",
    "email",
    "roles",
    "role_ids",
    "created_at",
    "last_modified",
    "enabled",
    "is_deleted"
FROM public.users_with_roles uwr
WHERE username = $1;

-- name: GetUserLogin :one
SELECT id, username, "password" , email, "enabled", "roles", "role_ids" FROM public.users_with_roles uwr
WHERE username = $1 OR email = $1
LIMIT 1;

-- name: GetUserIdByName :one
SELECT
	id
FROM public.users
where username = $1;

-- name: GetUserNameById :one
SELECT
  "username"
FROM public.users
WHERE id = $1;

-- name: UpdateUserPasswordById :exec
UPDATE users
  set password = $2
WHERE id = $1;

-- name: UpdateUserEmailById :one
UPDATE users
  set email = $2
WHERE id = $1
RETURNING *;

-- name: DeleteUserById :exec
DELETE FROM users
WHERE id = $1;

-- name: SoftDeleteUserById :one
UPDATE users
  set is_deleted = TRUE,
  "enabled" = FALSE
WHERE id = $1
RETURNING *;

-- name: DisableUserById :one
UPDATE users
  set "enabled" = $2
WHERE id = $1
RETURNING *;

-- name: EnableUserById :one
UPDATE users
  set "enabled" = $2
WHERE id = $1
RETURNING *;

-- name: InsertUserHostedDb :one
INSERT INTO public.user_hosted_db (
  price_tier_code_id,
  user_id,
  current_host_server_id,
  current_kube_cluster_id,
  user_application_ids,
  db_platform_id,
  fqdn,
  pub_ip_address,
  listen_port,
  private_ip_address,
  created_at,
  last_modified)
VALUES ($1, $2, $3, $4, $5, $6, $6, $7, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING *;

-- name: GetAllActiveUsers :many
SELECT
    "id",
    "username",
    "password",
     "email",
    "roles",
    "role_ids",
    "created_at",
    "last_modified",
    "enabled",
    "is_deleted"
FROM public.users_with_roles uwr;

-- name: GetAllUserPermissions :many
SELECT
  "UserId",
  "Username",
  "PermissionId",
  "Permission",
  "Role",
  "LastModified"
FROM
    public.user_permissions_view upv
ORDER BY "UserId" ASC;

-- name: GetUserPermissionsById :many
SELECT
  "UserId",
  "Username",
  "PermissionId",
  "Permission",
  "Role",
  "LastModified"
FROM
    public.user_permissions_view upv
WHERE "UserId" = $1;

-- name: VerifyUserPermissionById :one
SELECT EXISTS (
  SELECT
    "UserId",
    "Username",
    "PermissionId",
    "Permission",
    "Role",
    "LastModified"
  FROM
      public.user_permissions_view upv
  WHERE "UserId" = $1 and "Permission" = $2
);

-- name: VerifyUserPermissionByRoleId :one
SELECT EXISTS (
  SELECT
    "RoleId",
    "Role",
    "PermissionId",
    "Permission",
    "Role"
  FROM
      public.role_permissions_view rpv
  WHERE "RoleId" = $1 and "Permission" = $2
);

-- name: InsertOrUpdateUserRoleMappingById :one
INSERT INTO public.user_role_mapping(user_id, role_id, enabled)
VALUES ($1, $2, TRUE)
ON CONFLICT (user_id, role_id)
DO UPDATE SET enabled = TRUE
RETURNING *;

-- name: DisableUserRoleMappingById :one
UPDATE
  public.user_role_mapping
SET
  enabled = FALSE
WHERE user_id = $1 AND role_id = $2
RETURNING *;

-- name: GetRoleIdByName :one
SELECT
  "id" AS "RoleId"
FROM
  public. public.user_roles
WHERE "role_name" = $1;

-- name: InsertOrUpdateUserRole :one
INSERT INTO user_roles (id, role_name, role_description, created_at, last_modified, "enabled", "is_deleted")
VALUES(gen_random_uuid(), $1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, TRUE, false)
ON CONFLICT (role_name)
DO UPDATE SET
	role_description = EXCLUDED.role_description,
	last_modified = CURRENT_TIMESTAMP,
	"enabled" = TRUE,
	"is_deleted" = FALSE
RETURNING *;

-- name: EnableUserRoleById :exec
UPDATE user_roles SET "enabled" = TRUE
WHERE id = $1;

-- name: DisableUserRoleById :exec
UPDATE user_roles SET "enabled" = FALSE
WHERE id = $1;

-- name: SoftDeleteUserRoleById :exec
UPDATE user_roles
SET
"is_deleted" = TRUE,
"enabled" = FALSE
WHERE id = $1;

-- name: HardDeleteUserRoleById :exec
DELETE FROM user_roles
WHERE id = $1;

-- name: InsertOrUpdateAppPermission :one
INSERT INTO app_permissions(id, permission_name, permission_description)
VALUES(gen_random_uuid(), $1, $2)
ON CONFLICT (permission_name)
DO UPDATE SET
	permission_description = EXCLUDED.permission_description
RETURNING *;

-- name: InsertOrUpdateRolePermissionMapping :one
INSERT INTO role_permission_mapping(id, role_id, permission_id, "enabled", created_at, last_modified)
VALUES(gen_random_uuid(), $1, $2, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT(role_id, permission_id)
DO UPDATE SET
  role_id = EXCLUDED.role_id,
  permission_id = EXCLUDED.permission_id,
  "enabled" = true,
  created_at = CURRENT_TIMESTAMP,
  last_modified = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetAllUserRoles :many
SELECT "RoleId", "RoleName", "RoleDescription", "CreatedAt", "LastModified", "Enabled", "IsDeleted"
FROM public.user_roles_active;

-- name: GetAllAppPermissions :many
SELECT id, permission_name, permission_description
FROM public.app_permissions;

-- name: DbHealthCheckRead :one
SELECT id, status, check_type
FROM public.health_check WHERE check_type = 'Read'
LIMIT 1;


-- name: InsertExternalAuthToken :one
INSERT INTO public.external_auth_tokens (
    id,
    user_id,
    external_app_id,
    token,
    expiration
)
VALUES (
  gen_random_uuid(),  $1, $2, $3, $4
)
RETURNING id;

-- name: GetExternalAuthTokenById :one
SELECT
		id, user_id, external_app_id, token, expiration, created_at, last_modified
 FROM
  	public.external_auth_tokens WHERE id = $1;

-- name: GetExternalAuthTokensByUserIdAndAppId :many
SELECT
		id, user_id, external_app_id, token, expiration, created_at, last_modified
 FROM
  	public.external_auth_tokens 
WHERE user_id = $1 AND external_app_id = $2;

-- name: GetExternalAuthTokensByUserId :many
SELECT
		id, user_id, external_app_id, token, expiration, created_at, last_modified
 FROM
  	public.external_auth_tokens 
WHERE user_id = $1;

-- name: DeleteExternalAuthTokenById :exec
DELETE FROM external_auth_tokens
WHERE id = $1;

-- name: DeleteExpiredAuthTokens :exec
DELETE FROM external_auth_tokens
WHERE expiration < CURRENT_TIMESTAMP AT TIME ZONE 'UTC';

-- name: InsertExternalAppIntegrationByName :one
INSERT INTO public.external_integration_apps (id, "name") 
VALUES ($1, $2)
RETURNING *;

-- name: CreateExternalApplication :one
INSERT INTO public.external_integration_apps (id, "name", endpoint_url, app_description) 
VALUES ($1, $2, $3, $4)
RETURNING id, "name", created_at, last_modified, endpoint_url, app_description;

-- name: GetExternalApplicationById :one
SELECT id, "name", created_at, last_modified, endpoint_url, app_description
FROM public.external_integration_apps
WHERE id = $1;

-- name: GetExternalApplicationByName :one
SELECT id, "name", created_at, last_modified, endpoint_url, app_description
FROM public.external_integration_apps
WHERE "name" = $1;

-- name: UpdateExternalApplication :one
UPDATE public.external_integration_apps
SET 
    "name" = COALESCE($2, "name"),
    endpoint_url = COALESCE($3, endpoint_url),
    app_description = COALESCE($4, app_description),
    last_modified = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, "name", created_at, last_modified, endpoint_url, app_description;

-- name: DeleteExternalApplicationByName :exec
DELETE FROM external_integration_apps
WHERE "name" = $1;

-- name: DeleteExternalApplicationById :exec
DELETE FROM external_integration_apps
WHERE id = $1;

-- name: GetExternalAppIdByName :one
SELECT id FROM external_integration_apps WHERE "name" = $1;

-- name: GetExternalAppNameById :one
SELECT "name" FROM external_integration_apps WHERE id = $1;

-- name: GetAllExternalApps :many
SELECT id, "name" FROM external_integration_apps;

-- name: GetLatestExternalAuthToken :one
SELECT * FROM external_auth_tokens
WHERE user_id = $1 AND external_app_id = $2
ORDER BY created_at DESC
LIMIT 1;


-- name: GetLatestExternalAuthTokenByAppName :one
SELECT et.id, et.user_id, et.external_app_id, ea.name 
FROM external_auth_tokens et 
LEFT JOIN public.external_integration_apps ea on et.external_app_id = ea.id
WHERE et.user_id = $1 AND ea.name = $2
ORDER BY et.created_at DESC
LIMIT 1;

-- name: GetUserSecretsByUserId :many
SELECT
  auth_token_id,
  user_id,
  application_id,
  username,
  endpoint_url,
  email,
  application_name,
  token_created_at,
  expiration
FROM public.user_auth_app_mappings
WHERE user_id = $1;

-- name: GetUserSecretsByAppId :many
SELECT
  auth_token_id,
  user_id,
  application_id,
  username,
  endpoint_url,
  email,
  application_name,
  token_created_at,
  expiration
FROM public.user_auth_app_mappings
WHERE user_id = $1 AND application_id = $2;

-- name: GetUserSecretsByAppName :many
SELECT
  auth_token_id,
  user_id,
  application_id,
  username,
  endpoint_url,
  email,
  application_name,
  token_created_at,
  expiration
FROM public.user_auth_app_mappings
WHERE user_id = $1 AND application_name = $2;

-- SSH Keys CRUD Operations
-- name: CreateSSHKey :one
INSERT INTO public.ssh_keys (
    name,
    description,
    priv_secret_id,
    passphrase_id,
    public_key,
    key_type_id,
    owner_user_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING id, name, description, priv_secret_id, passphrase_id, public_key, key_type_id, owner_user_id, created_at, last_modified;

-- name: GetSSHKeyById :one
SELECT
    sk.id,
    sk.name,
    sk.description,
    sk.public_key,
    skt.name as key_type,
    sk.owner_user_id,
    sk.created_at,
    sk.last_modified,
    sk.priv_secret_id,
    sk.passphrase_id
FROM ssh_keys sk
JOIN ssh_key_types skt ON sk.key_type_id = skt.id
WHERE sk.id = $1;

-- name: GetSSHKeysByOwnerId :many
SELECT
    sk.id,
    sk.name,
    sk.description,
    sk.public_key,
    skt.name as key_type,
    sk.owner_user_id,
    sk.created_at,
    sk.last_modified,
    sk.priv_secret_id,
    sk.passphrase_id
FROM ssh_keys sk
JOIN ssh_key_types skt ON sk.key_type_id = skt.id
WHERE sk.owner_user_id = $1;

-- name: UpdateSSHKey :one
UPDATE public.ssh_keys
SET 
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    public_key = COALESCE($4, public_key),
    key_type_id = COALESCE($5, key_type_id),
    last_modified = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, name, description, priv_secret_id, public_key, key_type_id, owner_user_id, created_at, last_modified;

-- name: DeleteSSHKey :exec
DELETE FROM ssh_keys
WHERE id = $1;

-- SSH Key to Host Server Mappings CRUD Operations
-- name: CreateSSHKeyHostMapping :one
INSERT INTO public.host_server_ssh_mappings (
    host_server_id,
    ssh_key_id,
    user_id,
    hostserver_username,
    sudo_password_token_id
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (host_server_id, ssh_key_id) DO UPDATE SET
    hostserver_username = EXCLUDED.hostserver_username,
    user_id = EXCLUDED.user_id,
    sudo_password_token_id = EXCLUDED.sudo_password_token_id,
    last_modified = CURRENT_TIMESTAMP
RETURNING id, host_server_id, ssh_key_id, user_id, hostserver_username, sudo_password_token_id, created_at, last_modified;

-- name: GetSSHKeyHostMappingById :one
SELECT 
    id,
    host_server_id,
    ssh_key_id,
    user_id,
    hostserver_username,
    sudo_password_token_id,
    created_at,
    last_modified
FROM public.host_server_ssh_mappings
WHERE id = $1;

-- name: GetSSHKeyHostMappingsByUserId :many
SELECT 
    mapping_id,
    user_id,
    username,
    host_server_name,
    host_server_id,
    public_key,
    ssh_key_id,
    external_auth_token_id,
    ssh_key_type,
    hostserver_username,
    sudo_password_token_id,
    created_at,
    last_modified
FROM public.user_ssh_key_mappings
WHERE user_id = $1;

-- name: GetSSHKeyHostMappingsByHostId :many
SELECT 
    mapping_id,
    user_id,
    username,
    host_server_name,
    host_server_id,
    public_key,
    ssh_key_id,
    external_auth_token_id,
    ssh_key_type,
    hostserver_username,
    sudo_password_token_id,
    created_at,
    last_modified
FROM public.user_ssh_key_mappings
WHERE host_server_id = $1;

-- name: GetSSHKeyHostMappingsByKeyId :many
SELECT 
    mapping_id,
    user_id,
    username,
    host_server_name,
    host_server_id,
    public_key,
    ssh_key_id,
    external_auth_token_id,
    ssh_key_type,
    hostserver_username,
    sudo_password_token_id,
    created_at,
    last_modified
FROM public.user_ssh_key_mappings
WHERE ssh_key_id = $1;

-- name: UpdateSSHKeyHostMapping :one
UPDATE public.host_server_ssh_mappings
SET 
    hostserver_username = $2,
    last_modified = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, host_server_id, ssh_key_id, user_id, hostserver_username, sudo_password_token_id, created_at, last_modified;

-- name: DeleteSSHKeyHostMapping :exec
DELETE FROM public.host_server_ssh_mappings
WHERE id = $1;

-- name: DeleteSSHKeyHostMappingsBySshKeyId :exec
DELETE FROM public.host_server_ssh_mappings
WHERE ssh_key_id = $1;

-- Host Servers CRUD Operations
-- name: GetHostServerById :one
SELECT 
    id,
    hostname,
    ip_address,
    created_at,
    last_modified
FROM public.host_servers
WHERE id = $1;

-- name: GetHostServerByHostname :one
SELECT 
    id,
    hostname,
    ip_address,
    created_at,
    last_modified
FROM public.host_servers
WHERE hostname = $1;

-- name: GetHostServerByIP :one
SELECT 
    id,
    hostname,
    ip_address,
    created_at,
    last_modified
FROM public.host_servers
WHERE ip_address = $1;

-- name: GetAllHostServers :many
SELECT 
    id,
    hostname,
    ip_address,
    created_at,
    last_modified
FROM public.host_servers;

-- name: UpdateHostServer :one
UPDATE public.host_servers
SET 
    hostname = COALESCE($2, hostname),
    ip_address = COALESCE($3, ip_address),
    last_modified = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, hostname, ip_address, created_at, last_modified;

-- name: DeleteHostServer :exec
DELETE FROM public.host_servers
WHERE id = $1;

-- SSH Key Types Operations
-- name: GetAllSSHKeyTypes :many
SELECT 
    id,
    name,
    description,
    created_at,
    last_modified
FROM public.ssh_key_types;

-- name: GetSSHKeyTypeByName :one
SELECT 
    id,
    name,
    description,
    created_at,
    last_modified
FROM public.ssh_key_types
WHERE name = $1;

-- name: CreateSSHKeyType :one
INSERT INTO public.ssh_key_types (
    name,
    description
) VALUES (
    $1, $2
) RETURNING id, name, description, created_at, last_modified;

-- name: UpdateSSHKeyType :one
UPDATE public.ssh_key_types
SET 
    description = COALESCE($2, description),
    last_modified = CURRENT_TIMESTAMP
WHERE name = $1
RETURNING id, name, description, created_at, last_modified;

-- name: DeleteSSHKeyType :exec
DELETE FROM public.ssh_key_types
WHERE name = $1;

-- Host Servers CRUD Operations
-- name: CreateHostServer :one
INSERT INTO public.host_servers (
    hostname,
    ip_address
) VALUES (
    $1, $2
)
ON CONFLICT (hostname, ip_address) DO UPDATE SET
    last_modified = CURRENT_TIMESTAMP
RETURNING id, hostname, ip_address, created_at, last_modified;

-- name: CreateSSHSession :exec
INSERT INTO ssh_sessions (id, user_id, host_server_id, username, created_at, last_activity, is_active)
VALUES ($1, $2, $3, $4, $5, $6, true)
ON CONFLICT (id) DO UPDATE SET last_activity = $6, is_active = true;

-- name: GetSSHSessionById :one
SELECT id, user_id, host_server_id, username, created_at, last_activity
FROM ssh_sessions WHERE id = $1 AND is_active = true;

-- name: ListActiveSSHSessions :many
SELECT id, user_id, host_server_id, username, created_at, last_activity
FROM ssh_sessions WHERE is_active = true;

-- name: RemoveSSHSession :exec
UPDATE ssh_sessions SET is_active = false WHERE id = $1;

-- name: UpdateSSHSessionActivity :exec
UPDATE ssh_sessions SET last_activity = $2 WHERE id = $1;

-- name: MarkSSHSessionInactive :exec
UPDATE ssh_sessions SET is_active = false WHERE id = $1;

-- Host Server Types CRUD Operations
-- name: CreateHostServerType :one
INSERT INTO public.host_server_types (name) VALUES ($1)
ON CONFLICT (name) DO UPDATE SET last_modified = CURRENT_TIMESTAMP
RETURNING host_server_type_id, name, last_modified;

-- name: GetHostServerTypeById :one
SELECT host_server_type_id, name, last_modified
FROM public.host_server_types
WHERE host_server_type_id = $1;

-- name: GetHostServerTypeByName :one
SELECT host_server_type_id, name, last_modified
FROM public.host_server_types
WHERE name = $1;

-- name: GetAllHostServerTypes :many
SELECT host_server_type_id, name, last_modified
FROM public.host_server_types
ORDER BY name;

-- name: UpdateHostServerType :one
UPDATE public.host_server_types
SET name = $2, last_modified = CURRENT_TIMESTAMP
WHERE host_server_type_id = $1
RETURNING host_server_type_id, name, last_modified;

-- name: DeleteHostServerType :exec
DELETE FROM public.host_server_types
WHERE host_server_type_id = $1;

-- Platform Types CRUD Operations
-- name: CreatePlatformType :one
INSERT INTO public.platform_types (name) VALUES ($1)
ON CONFLICT (name) DO UPDATE SET last_modified = CURRENT_TIMESTAMP
RETURNING platform_type_id, name, last_modified;

-- name: GetPlatformTypeById :one
SELECT platform_type_id, name, last_modified
FROM public.platform_types
WHERE platform_type_id = $1;

-- name: GetPlatformTypeByName :one
SELECT platform_type_id, name, last_modified
FROM public.platform_types
WHERE name = $1;

-- name: GetAllPlatformTypes :many
SELECT platform_type_id, name, last_modified
FROM public.platform_types
ORDER BY name;

-- name: UpdatePlatformType :one
UPDATE public.platform_types
SET name = $2, last_modified = CURRENT_TIMESTAMP
WHERE platform_type_id = $1
RETURNING platform_type_id, name, last_modified;

-- name: DeletePlatformType :exec
DELETE FROM public.platform_types
WHERE platform_type_id = $1;

-- Host Server Type Mappings CRUD Operations
-- name: CreateHostServerTypeMapping :one
INSERT INTO public.host_server_type_mappings (host_server_id, host_server_type_id)
VALUES ($1, $2)
ON CONFLICT (host_server_id, host_server_type_id) DO UPDATE SET last_modified = CURRENT_TIMESTAMP
RETURNING id, host_server_id, host_server_type_id, created_at, last_modified;

-- name: GetHostServerTypeMappingsByHostId :many
SELECT 
    hstm.id,
    hstm.host_server_id,
    hstm.host_server_type_id,
    hst.name as host_server_type_name,
    hstm.created_at,
    hstm.last_modified
FROM public.host_server_type_mappings hstm
JOIN public.host_server_types hst ON hstm.host_server_type_id = hst.host_server_type_id
WHERE hstm.host_server_id = $1;

-- name: GetHostServerTypeMappingsByTypeId :many
SELECT 
    hstm.id,
    hstm.host_server_id,
    hstm.host_server_type_id,
    hs.hostname,
    hs.ip_address,
    hstm.created_at,
    hstm.last_modified
FROM public.host_server_type_mappings hstm
JOIN public.host_servers hs ON hstm.host_server_id = hs.id
WHERE hstm.host_server_type_id = $1;

-- name: DeleteHostServerTypeMapping :exec
DELETE FROM public.host_server_type_mappings
WHERE id = $1;

-- name: DeleteHostServerTypeMappingsByHostId :exec
DELETE FROM public.host_server_type_mappings
WHERE host_server_id = $1;

-- Platform Type Mappings CRUD Operations
-- name: CreatePlatformTypeMapping :one
INSERT INTO public.platform_type_mappings (platform_type_id, host_server_id, host_server_type_id)
VALUES ($1, $2, $3)
ON CONFLICT (platform_type_id, host_server_id, host_server_type_id) DO UPDATE SET last_modified = CURRENT_TIMESTAMP
RETURNING id, platform_type_id, host_server_id, host_server_type_id, created_at, last_modified;

-- name: GetPlatformTypeMappingsByHostId :many
SELECT 
    ptm.id,
    ptm.platform_type_id,
    ptm.host_server_id,
    ptm.host_server_type_id,
    pt.name as platform_type_name,
    hst.name as host_server_type_name,
    ptm.created_at,
    ptm.last_modified
FROM public.platform_type_mappings ptm
JOIN public.platform_types pt ON ptm.platform_type_id = pt.platform_type_id
JOIN public.host_server_types hst ON ptm.host_server_type_id = hst.host_server_type_id
WHERE ptm.host_server_id = $1;

-- name: GetPlatformTypeMappingsByPlatformId :many
SELECT 
    ptm.id,
    ptm.platform_type_id,
    ptm.host_server_id,
    ptm.host_server_type_id,
    hs.hostname,
    hs.ip_address,
    hst.name as host_server_type_name,
    ptm.created_at,
    ptm.last_modified
FROM public.platform_type_mappings ptm
JOIN public.host_servers hs ON ptm.host_server_id = hs.id
JOIN public.host_server_types hst ON ptm.host_server_type_id = hst.host_server_type_id
WHERE ptm.platform_type_id = $1;

-- name: GetPlatformTypeMappingsByHostServerTypeId :many
SELECT 
    ptm.id,
    ptm.platform_type_id,
    ptm.host_server_id,
    ptm.host_server_type_id,
    pt.name as platform_type_name,
    hs.hostname,
    hs.ip_address,
    ptm.created_at,
    ptm.last_modified
FROM public.platform_type_mappings ptm
JOIN public.platform_types pt ON ptm.platform_type_id = pt.platform_type_id
JOIN public.host_servers hs ON ptm.host_server_id = hs.id
WHERE ptm.host_server_type_id = $1;

-- name: DeletePlatformTypeMapping :exec
DELETE FROM public.platform_type_mappings
WHERE id = $1;

-- name: DeletePlatformTypeMappingsByHostId :exec
DELETE FROM public.platform_type_mappings
WHERE host_server_id = $1;
