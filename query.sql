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

-- name: InsertHostServer :one
INSERT INTO host_servers (
            hostname, ip_address, is_container_host, is_vm_host, is_virtual_machine, id_db_host
        ) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (hostname, ip_address)
        DO UPDATE SET
            is_container_host = EXCLUDED.is_container_host,
            is_vm_host = EXCLUDED.is_vm_host,
            is_virtual_machine = EXCLUDED.is_virtual_machine,
            id_db_host = EXCLUDED.id_db_host,
			      last_modified = DEFAULT
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


-- name: InsertExternalAuthToken :exec
INSERT INTO public.external_auth_tokens (
    id,
    user_id,
    external_app_id,
    token,
    expiration
)
VALUES (
  gen_random_uuid(),  $1, $2, $3, $4
);

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
    public_key,
    key_type_id,
    owner_user_id
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING id, name, description, priv_secret_id, public_key, key_type_id, owner_user_id, created_at, last_modified;

-- name: GetSSHKeyById :one
SELECT 
    sk.id,
    sk.name,
    sk.description,
    sk.priv_secret_id,
    sk.public_key,
    sk.key_type_id,
    sk.owner_user_id,
    sk.created_at,
    sk.last_modified,
    skt.name as key_type_name,
    skt.description as key_type_description,
    u.username as owner_username
FROM public.ssh_keys sk
JOIN public.ssh_key_types skt ON sk.key_type_id = skt.id
JOIN public.users u ON sk.owner_user_id = u.id
WHERE sk.id = $1;

-- name: GetSSHKeysByOwnerId :many
SELECT 
    sk.id,
    sk.name,
    sk.description,
    sk.priv_secret_id,
    sk.public_key,
    sk.key_type_id,
    sk.owner_user_id,
    sk.created_at,
    sk.last_modified,
    skt.name as key_type_name,
    skt.description as key_type_description,
    u.username as owner_username
FROM public.ssh_keys sk
JOIN public.ssh_key_types skt ON sk.key_type_id = skt.id
JOIN public.users u ON sk.owner_user_id = u.id
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
DELETE FROM public.ssh_keys
WHERE id = $1;

-- SSH Key to Host Server Mappings CRUD Operations
-- name: CreateSSHKeyHostMapping :one
INSERT INTO public.ssh_key_host_mappings (
    ssh_key_id,
    host_server_id,
    user_id,
    hostserver_username
) VALUES (
    $1, $2, $3, $4
) RETURNING id, ssh_key_id, host_server_id, user_id, hostserver_username, created_at, last_modified;

-- name: GetSSHKeyHostMappingById :one
SELECT 
    id,
    ssh_key_id,
    host_server_id,
    user_id,
    hostserver_username,
    created_at,
    last_modified
FROM public.ssh_key_host_mappings
WHERE id = $1;

-- name: GetSSHKeyHostMappingsByUserId :many
SELECT 
    user_id,
    username,
    host_server_name,
    host_server_id,
    public_key,
    ssh_key_id,
    external_auth_token_id,
    ssh_key_type,
    hostserver_username
FROM public.user_ssh_key_mappings
WHERE user_id = $1;

-- name: GetSSHKeyHostMappingsByHostId :many
SELECT 
    user_id,
    username,
    host_server_name,
    host_server_id,
    public_key,
    ssh_key_id,
    external_auth_token_id,
    ssh_key_type,
    hostserver_username
FROM public.user_ssh_key_mappings
WHERE host_server_id = $1;

-- name: GetSSHKeyHostMappingsByKeyId :many
SELECT 
    user_id,
    username,
    host_server_name,
    host_server_id,
    public_key,
    ssh_key_id,
    external_auth_token_id,
    ssh_key_type,
    hostserver_username
FROM public.user_ssh_key_mappings
WHERE ssh_key_id = $1;

-- name: UpdateSSHKeyHostMapping :one
UPDATE public.ssh_key_host_mappings
SET 
    hostserver_username = COALESCE($2, hostserver_username),
    last_modified = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, ssh_key_id, host_server_id, user_id, hostserver_username, created_at, last_modified;

-- name: DeleteSSHKeyHostMapping :exec
DELETE FROM public.ssh_key_host_mappings
WHERE id = $1;

-- name: DeleteSSHKeyHostMappingsBySshKeyId :exec
DELETE FROM public.ssh_key_host_mappings
WHERE ssh_key_id = $1;

-- Host Servers CRUD Operations
-- name: CreateHostServer :one
INSERT INTO public.host_servers (
    hostname,
    ip_address,
    is_container_host,
    is_vm_host,
    is_virtual_machine,
    id_db_host
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING id, hostname, ip_address, is_container_host, is_vm_host, is_virtual_machine, id_db_host, created_at, last_modified;

-- name: GetHostServerById :one
SELECT 
    id,
    hostname,
    ip_address,
    is_container_host,
    is_vm_host,
    is_virtual_machine,
    id_db_host,
    created_at,
    last_modified
FROM public.host_servers
WHERE id = $1;

-- name: GetHostServerByHostname :one
SELECT 
    id,
    hostname,
    ip_address,
    is_container_host,
    is_vm_host,
    is_virtual_machine,
    id_db_host,
    created_at,
    last_modified
FROM public.host_servers
WHERE hostname = $1;

-- name: GetHostServerByIP :one
SELECT 
    id,
    hostname,
    ip_address,
    is_container_host,
    is_vm_host,
    is_virtual_machine,
    id_db_host,
    created_at,
    last_modified
FROM public.host_servers
WHERE ip_address = $1;

-- name: GetAllHostServers :many
SELECT 
    id,
    hostname,
    ip_address,
    is_container_host,
    is_vm_host,
    is_virtual_machine,
    id_db_host,
    created_at,
    last_modified
FROM public.host_servers;

-- name: UpdateHostServer :one
UPDATE public.host_servers
SET 
    hostname = COALESCE($2, hostname),
    ip_address = COALESCE($3, ip_address),
    is_container_host = COALESCE($4, is_container_host),
    is_vm_host = COALESCE($5, is_vm_host),
    is_virtual_machine = COALESCE($6, is_virtual_machine),
    id_db_host = COALESCE($7, id_db_host),
    last_modified = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, hostname, ip_address, is_container_host, is_vm_host, is_virtual_machine, id_db_host, created_at, last_modified;

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
