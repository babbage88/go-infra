-- name: CreateUser :one
INSERT INTO users (
    username,
    password,
    email,
    role
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetUserById :one
SELECT
	id,
	username,
	"password",
	email,
	"role",
	created_at,
	last_modified,
	"enabled",
	is_deleted
FROM public.users
WHERE id = $1;

-- name: GetUserByName :one
SELECT
	id,
	username,
	"password",
	email,
	"role",
	created_at,
	last_modified,
	"enabled",
	is_deleted
FROM public.users
WHERE username = $1;

-- name: GetUserLogin :one
SELECT id, username, "password" , email, "enabled", "role" FROM public.users
WHERE username = $1
LIMIT 1;

-- name: GetUserIdByName :one
SELECT
	id
FROM public.users
where username = $1;

-- name: UpdateUserPasswordById :exec
UPDATE users
  set password = $2
WHERE id = $1;

-- name: UpdateUserEmailById :one
UPDATE users
  set email = $2
WHERE id = $1
RETURNING *;

-- name: UpdateUserRoleById :one
UPDATE users
  set email = $2
WHERE id = $1
RETURNING *;

-- name: DeleteUserById :exec
DELETE FROM users
WHERE id = $1;

-- name: SoftDeleteUserById :one
UPDATE users
  set is_deleted = $2
WHERE id = $1
RETURNING *;

-- name: DisableUserById :one
UPDATE users
  set "enabled" = $2
WHERE id = $1
RETURNING *;

-- name: InsertHostServer :one
INSERT INTO host_servers (
            hostname, ip_address, username, public_ssh_keyname, hosted_domains,
            ssl_key_path, is_container_host, is_vm_host, is_virtual_machine, id_db_host
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (hostname, ip_address)
        DO UPDATE SET
            username = EXCLUDED.username,
            public_ssh_keyname = EXCLUDED.public_ssh_keyname,
            hosted_domains = EXCLUDED.hosted_domains,
            ssl_key_path = EXCLUDED.ssl_key_path,
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

-- name: InsertAuthToken :exec
INSERT INTO auth_tokens (user_id, token, expiration)
VALUES ($1, $2, $3);

-- name: GetAuthTokenFromDb :one
SELECT
		id, user_id, token, expiration, created_at, last_modified
 FROM
  	public.auth_tokens WHERE id = $1;

-- name: DeleteAuthTokenById :exec
DELETE FROM auth_tokens
WHERE id = $1;

-- name: DeleteExpiredAuthTokens :exec
DELETE FROM auth_tokens
WHERE expiration < CURRENT_TIMESTAMP AT TIME ZONE 'UTC';

-- name: GetAllActiveUsers :many
SELECT 	id,
	username,
	email,
	"role",
	created_at,
	last_modified,
	"enabled"
FROM users
where "enabled" is TRUE;

-- name: GetAllUserPermissions :many
SELECT
  "Permission",
  "Username",
  "UserId",
  "PermissionId",
  "Role",
  "LastModified"
FROM public.user_permissions_view
ORDER BY "UserId" ASC;

-- name: GetUserPermissionsById :many
SELECT
  "Permission",
  "Username",
  "UserId",
  "PermissionId",
  "Role",
  "LastModified"
FROM public.user_permissions_view
WHERE "UserId" = $1;

-- name: VerifyUserPermissionById :one
SELECT
  "Permission",
  "Username",
  "UserId",
  "PermissionId",
  "Role",
  "LastModified"
FROM public.user_permissions_view
WHERE "UserId" = $1 and "Permission" = $2;
