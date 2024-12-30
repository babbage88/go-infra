// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package infra_db_pg

import (
	"context"
	"net/netip"

	"github.com/jackc/pgx/v5/pgtype"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (
    username,
    password,
    email
) VALUES (
  $1, $2, $3
)
RETURNING id, username, password, email, created_at, last_modified, enabled, is_deleted
`

type CreateUserParams struct {
	Username pgtype.Text
	Password pgtype.Text
	Email    pgtype.Text
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Username, arg.Password, arg.Email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.Email,
		&i.CreatedAt,
		&i.LastModified,
		&i.Enabled,
		&i.IsDeleted,
	)
	return i, err
}

const deleteAuthTokenById = `-- name: DeleteAuthTokenById :exec
DELETE FROM auth_tokens
WHERE id = $1
`

func (q *Queries) DeleteAuthTokenById(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteAuthTokenById, id)
	return err
}

const deleteExpiredAuthTokens = `-- name: DeleteExpiredAuthTokens :exec
DELETE FROM auth_tokens
WHERE expiration < CURRENT_TIMESTAMP AT TIME ZONE 'UTC'
`

func (q *Queries) DeleteExpiredAuthTokens(ctx context.Context) error {
	_, err := q.db.Exec(ctx, deleteExpiredAuthTokens)
	return err
}

const deleteUserById = `-- name: DeleteUserById :exec
DELETE FROM users
WHERE id = $1
`

func (q *Queries) DeleteUserById(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, deleteUserById, id)
	return err
}

const disableUserById = `-- name: DisableUserById :one
UPDATE users
  set "enabled" = $2
WHERE id = $1
RETURNING id, username, password, email, created_at, last_modified, enabled, is_deleted
`

type DisableUserByIdParams struct {
	ID      int32
	Enabled bool
}

func (q *Queries) DisableUserById(ctx context.Context, arg DisableUserByIdParams) (User, error) {
	row := q.db.QueryRow(ctx, disableUserById, arg.ID, arg.Enabled)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.Email,
		&i.CreatedAt,
		&i.LastModified,
		&i.Enabled,
		&i.IsDeleted,
	)
	return i, err
}

const disableUserRoleById = `-- name: DisableUserRoleById :exec
UPDATE user_roles SET "enabled" = FALSE
WHERE id = $1
`

func (q *Queries) DisableUserRoleById(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, disableUserRoleById, id)
	return err
}

const disableUserRoleMappingById = `-- name: DisableUserRoleMappingById :one
UPDATE
  public.user_role_mapping
SET
  enabled = FALSE
WHERE user_id = $1 AND role_id = $2
RETURNING id, user_id, role_id, enabled, created_at, last_modified
`

type DisableUserRoleMappingByIdParams struct {
	UserID int32
	RoleID int32
}

func (q *Queries) DisableUserRoleMappingById(ctx context.Context, arg DisableUserRoleMappingByIdParams) (UserRoleMapping, error) {
	row := q.db.QueryRow(ctx, disableUserRoleMappingById, arg.UserID, arg.RoleID)
	var i UserRoleMapping
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.RoleID,
		&i.Enabled,
		&i.CreatedAt,
		&i.LastModified,
	)
	return i, err
}

const enableUserById = `-- name: EnableUserById :one
UPDATE users
  set "enabled" = $2
WHERE id = $1
RETURNING id, username, password, email, created_at, last_modified, enabled, is_deleted
`

type EnableUserByIdParams struct {
	ID      int32
	Enabled bool
}

func (q *Queries) EnableUserById(ctx context.Context, arg EnableUserByIdParams) (User, error) {
	row := q.db.QueryRow(ctx, enableUserById, arg.ID, arg.Enabled)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.Email,
		&i.CreatedAt,
		&i.LastModified,
		&i.Enabled,
		&i.IsDeleted,
	)
	return i, err
}

const enableUserRoleById = `-- name: EnableUserRoleById :exec
UPDATE user_roles SET "enabled" = TRUE
WHERE id = $1
`

func (q *Queries) EnableUserRoleById(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, enableUserRoleById, id)
	return err
}

const getAllActiveUsers = `-- name: GetAllActiveUsers :many
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
`

func (q *Queries) GetAllActiveUsers(ctx context.Context) ([]UsersWithRole, error) {
	rows, err := q.db.Query(ctx, getAllActiveUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []UsersWithRole
	for rows.Next() {
		var i UsersWithRole
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Password,
			&i.Email,
			&i.Roles,
			&i.RoleIds,
			&i.CreatedAt,
			&i.LastModified,
			&i.Enabled,
			&i.IsDeleted,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllAppPermissions = `-- name: GetAllAppPermissions :many
SELECT id, permission_name, permission_description
FROM public.app_permissions
`

func (q *Queries) GetAllAppPermissions(ctx context.Context) ([]AppPermission, error) {
	rows, err := q.db.Query(ctx, getAllAppPermissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AppPermission
	for rows.Next() {
		var i AppPermission
		if err := rows.Scan(&i.ID, &i.PermissionName, &i.PermissionDescription); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllUserRoles = `-- name: GetAllUserRoles :many
SELECT "RoleId", "RoleName", "RoleDescription", "CreatedAt", "LastModified", "Enabled", "IsDeleted"
FROM public.user_roles_active
`

func (q *Queries) GetAllUserRoles(ctx context.Context) ([]UserRolesActive, error) {
	rows, err := q.db.Query(ctx, getAllUserRoles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []UserRolesActive
	for rows.Next() {
		var i UserRolesActive
		if err := rows.Scan(
			&i.RoleId,
			&i.RoleName,
			&i.RoleDescription,
			&i.CreatedAt,
			&i.LastModified,
			&i.Enabled,
			&i.IsDeleted,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAuthTokenFromDb = `-- name: GetAuthTokenFromDb :one
SELECT
		id, user_id, token, expiration, created_at, last_modified
 FROM
  	public.auth_tokens WHERE id = $1
`

func (q *Queries) GetAuthTokenFromDb(ctx context.Context, id int32) (AuthToken, error) {
	row := q.db.QueryRow(ctx, getAuthTokenFromDb, id)
	var i AuthToken
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Token,
		&i.Expiration,
		&i.CreatedAt,
		&i.LastModified,
	)
	return i, err
}

const getRoleIdByName = `-- name: GetRoleIdByName :one
SELECT
  "id" AS "RoleId"
FROM
  public. public.user_roles
WHERE "role_name" = $1
`

func (q *Queries) GetRoleIdByName(ctx context.Context, roleName string) (int32, error) {
	row := q.db.QueryRow(ctx, getRoleIdByName, roleName)
	var RoleId int32
	err := row.Scan(&RoleId)
	return RoleId, err
}

const getUserById = `-- name: GetUserById :one
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
WHERE "id" = $1
`

func (q *Queries) GetUserById(ctx context.Context, id int32) (UsersWithRole, error) {
	row := q.db.QueryRow(ctx, getUserById, id)
	var i UsersWithRole
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.Email,
		&i.Roles,
		&i.RoleIds,
		&i.CreatedAt,
		&i.LastModified,
		&i.Enabled,
		&i.IsDeleted,
	)
	return i, err
}

const getUserByName = `-- name: GetUserByName :one
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
WHERE username = $1
`

func (q *Queries) GetUserByName(ctx context.Context, username pgtype.Text) (UsersWithRole, error) {
	row := q.db.QueryRow(ctx, getUserByName, username)
	var i UsersWithRole
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.Email,
		&i.Roles,
		&i.RoleIds,
		&i.CreatedAt,
		&i.LastModified,
		&i.Enabled,
		&i.IsDeleted,
	)
	return i, err
}

const getUserIdByName = `-- name: GetUserIdByName :one
SELECT
	id
FROM public.users
where username = $1
`

func (q *Queries) GetUserIdByName(ctx context.Context, username pgtype.Text) (int32, error) {
	row := q.db.QueryRow(ctx, getUserIdByName, username)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const getUserLogin = `-- name: GetUserLogin :one
SELECT id, username, "password" , email, "enabled", "roles", "role_ids" FROM public.users_with_roles uwr
WHERE username = $1
LIMIT 1
`

type GetUserLoginRow struct {
	ID       int32
	Username pgtype.Text
	Password pgtype.Text
	Email    pgtype.Text
	Enabled  bool
	Roles    interface{}
	RoleIds  interface{}
}

func (q *Queries) GetUserLogin(ctx context.Context, username pgtype.Text) (GetUserLoginRow, error) {
	row := q.db.QueryRow(ctx, getUserLogin, username)
	var i GetUserLoginRow
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.Email,
		&i.Enabled,
		&i.Roles,
		&i.RoleIds,
	)
	return i, err
}

const getUserPermissionsById = `-- name: GetUserPermissionsById :many
SELECT
  "UserId",
  "Username",
  "PermissionId",
  "Permission",
  "Role",
  "LastModified"
FROM
    public.user_permissions_view upv
WHERE "UserId" = $1
`

func (q *Queries) GetUserPermissionsById(ctx context.Context, userid pgtype.Int4) ([]UserPermissionsView, error) {
	rows, err := q.db.Query(ctx, getUserPermissionsById, userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []UserPermissionsView
	for rows.Next() {
		var i UserPermissionsView
		if err := rows.Scan(
			&i.UserId,
			&i.Username,
			&i.PermissionId,
			&i.Permission,
			&i.Role,
			&i.LastModified,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const hardDeleteUserRoleById = `-- name: HardDeleteUserRoleById :exec
DELETE FROM user_roles
WHERE id = $1
`

func (q *Queries) HardDeleteUserRoleById(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, hardDeleteUserRoleById, id)
	return err
}

const insertAuthToken = `-- name: InsertAuthToken :exec
INSERT INTO auth_tokens (user_id, token, expiration)
VALUES ($1, $2, $3)
`

type InsertAuthTokenParams struct {
	UserID     pgtype.Int4
	Token      pgtype.Text
	Expiration pgtype.Timestamp
}

func (q *Queries) InsertAuthToken(ctx context.Context, arg InsertAuthTokenParams) error {
	_, err := q.db.Exec(ctx, insertAuthToken, arg.UserID, arg.Token, arg.Expiration)
	return err
}

const insertHostServer = `-- name: InsertHostServer :one
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
RETURNING id, hostname, ip_address, username, public_ssh_keyname, hosted_domains, ssl_key_path, is_container_host, is_vm_host, is_virtual_machine, id_db_host, created_at, last_modified
`

type InsertHostServerParams struct {
	Hostname         string
	IpAddress        netip.Addr
	Username         pgtype.Text
	PublicSshKeyname pgtype.Text
	HostedDomains    []string
	SslKeyPath       pgtype.Text
	IsContainerHost  pgtype.Bool
	IsVmHost         pgtype.Bool
	IsVirtualMachine pgtype.Bool
	IDDbHost         pgtype.Bool
}

func (q *Queries) InsertHostServer(ctx context.Context, arg InsertHostServerParams) (HostServer, error) {
	row := q.db.QueryRow(ctx, insertHostServer,
		arg.Hostname,
		arg.IpAddress,
		arg.Username,
		arg.PublicSshKeyname,
		arg.HostedDomains,
		arg.SslKeyPath,
		arg.IsContainerHost,
		arg.IsVmHost,
		arg.IsVirtualMachine,
		arg.IDDbHost,
	)
	var i HostServer
	err := row.Scan(
		&i.ID,
		&i.Hostname,
		&i.IpAddress,
		&i.Username,
		&i.PublicSshKeyname,
		&i.HostedDomains,
		&i.SslKeyPath,
		&i.IsContainerHost,
		&i.IsVmHost,
		&i.IsVirtualMachine,
		&i.IDDbHost,
		&i.CreatedAt,
		&i.LastModified,
	)
	return i, err
}

const insertOrUpdateAppPermission = `-- name: InsertOrUpdateAppPermission :one
INSERT INTO app_permissions(id, permission_name, permission_description)
VALUES(nextval('app_permissions_id_seq'::regclass), $1, $2)
ON CONFLICT (permission_name)
DO UPDATE SET
	permission_description = EXCLUDED.permission_description
RETURNING id, permission_name, permission_description
`

type InsertOrUpdateAppPermissionParams struct {
	PermissionName        string
	PermissionDescription pgtype.Text
}

func (q *Queries) InsertOrUpdateAppPermission(ctx context.Context, arg InsertOrUpdateAppPermissionParams) (AppPermission, error) {
	row := q.db.QueryRow(ctx, insertOrUpdateAppPermission, arg.PermissionName, arg.PermissionDescription)
	var i AppPermission
	err := row.Scan(&i.ID, &i.PermissionName, &i.PermissionDescription)
	return i, err
}

const insertOrUpdateRolePermissionMapping = `-- name: InsertOrUpdateRolePermissionMapping :one
INSERT INTO role_permission_mapping(id, role_id, permission_id, "enabled", created_at, last_modified)
VALUES(nextval('role_permission_mapping_id_seq'::regclass), $1, $2, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT(role_id, permission_id)
DO UPDATE SET
  role_id = EXCLUDED.role_id,
  permission_id = EXCLUDED.permission_id,
  "enabled" = true,
  created_at = CURRENT_TIMESTAMP,
  last_modified = CURRENT_TIMESTAMP
RETURNING id, role_id, permission_id, enabled, created_at, last_modified
`

type InsertOrUpdateRolePermissionMappingParams struct {
	RoleID       int32
	PermissionID int32
}

func (q *Queries) InsertOrUpdateRolePermissionMapping(ctx context.Context, arg InsertOrUpdateRolePermissionMappingParams) (RolePermissionMapping, error) {
	row := q.db.QueryRow(ctx, insertOrUpdateRolePermissionMapping, arg.RoleID, arg.PermissionID)
	var i RolePermissionMapping
	err := row.Scan(
		&i.ID,
		&i.RoleID,
		&i.PermissionID,
		&i.Enabled,
		&i.CreatedAt,
		&i.LastModified,
	)
	return i, err
}

const insertOrUpdateUserRole = `-- name: InsertOrUpdateUserRole :one
INSERT INTO user_roles (id, role_name, role_description, created_at, last_modified, "enabled", "is_deleted")
VALUES(nextval('user_roles_id_seq'::regclass), $1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, TRUE, false)
ON CONFLICT (role_name)
DO UPDATE SET
	role_description = EXCLUDED.role_description,
	last_modified = CURRENT_TIMESTAMP,
	"enabled" = TRUE,
	"is_deleted" = FALSE
RETURNING id, role_name, role_description, created_at, last_modified, enabled, is_deleted
`

type InsertOrUpdateUserRoleParams struct {
	RoleName        string
	RoleDescription pgtype.Text
}

func (q *Queries) InsertOrUpdateUserRole(ctx context.Context, arg InsertOrUpdateUserRoleParams) (UserRole, error) {
	row := q.db.QueryRow(ctx, insertOrUpdateUserRole, arg.RoleName, arg.RoleDescription)
	var i UserRole
	err := row.Scan(
		&i.ID,
		&i.RoleName,
		&i.RoleDescription,
		&i.CreatedAt,
		&i.LastModified,
		&i.Enabled,
		&i.IsDeleted,
	)
	return i, err
}

const insertOrUpdateUserRoleMappingById = `-- name: InsertOrUpdateUserRoleMappingById :one
INSERT INTO public.user_role_mapping(user_id, role_id, enabled)
VALUES ($1, $2, TRUE)
ON CONFLICT (user_id, role_id)
DO UPDATE SET enabled = TRUE
RETURNING id, user_id, role_id, enabled, created_at, last_modified
`

type InsertOrUpdateUserRoleMappingByIdParams struct {
	UserID int32
	RoleID int32
}

func (q *Queries) InsertOrUpdateUserRoleMappingById(ctx context.Context, arg InsertOrUpdateUserRoleMappingByIdParams) (UserRoleMapping, error) {
	row := q.db.QueryRow(ctx, insertOrUpdateUserRoleMappingById, arg.UserID, arg.RoleID)
	var i UserRoleMapping
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.RoleID,
		&i.Enabled,
		&i.CreatedAt,
		&i.LastModified,
	)
	return i, err
}

const insertUserHostedDb = `-- name: InsertUserHostedDb :one
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
RETURNING id, price_tier_code_id, user_id, current_host_server_id, current_kube_cluster_id, user_application_ids, db_platform_id, fqdn, pub_ip_address, listen_port, private_ip_address, created_at, last_modified
`

type InsertUserHostedDbParams struct {
	PriceTierCodeID      int32
	UserID               int32
	CurrentHostServerID  int32
	CurrentKubeClusterID pgtype.Int4
	UserApplicationIds   []int32
	DbPlatformID         int32
	PubIpAddress         netip.Addr
	PrivateIpAddress     *netip.Addr
}

func (q *Queries) InsertUserHostedDb(ctx context.Context, arg InsertUserHostedDbParams) (UserHostedDb, error) {
	row := q.db.QueryRow(ctx, insertUserHostedDb,
		arg.PriceTierCodeID,
		arg.UserID,
		arg.CurrentHostServerID,
		arg.CurrentKubeClusterID,
		arg.UserApplicationIds,
		arg.DbPlatformID,
		arg.PubIpAddress,
		arg.PrivateIpAddress,
	)
	var i UserHostedDb
	err := row.Scan(
		&i.ID,
		&i.PriceTierCodeID,
		&i.UserID,
		&i.CurrentHostServerID,
		&i.CurrentKubeClusterID,
		&i.UserApplicationIds,
		&i.DbPlatformID,
		&i.Fqdn,
		&i.PubIpAddress,
		&i.ListenPort,
		&i.PrivateIpAddress,
		&i.CreatedAt,
		&i.LastModified,
	)
	return i, err
}

const softDeleteUserById = `-- name: SoftDeleteUserById :one
UPDATE users
  set is_deleted = TRUE,
  "enabled" = FALSE
WHERE id = $1
RETURNING id, username, password, email, created_at, last_modified, enabled, is_deleted
`

func (q *Queries) SoftDeleteUserById(ctx context.Context, id int32) (User, error) {
	row := q.db.QueryRow(ctx, softDeleteUserById, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.Email,
		&i.CreatedAt,
		&i.LastModified,
		&i.Enabled,
		&i.IsDeleted,
	)
	return i, err
}

const softDeleteUserRoleById = `-- name: SoftDeleteUserRoleById :exec
UPDATE user_roles
SET
"is_deleted" = TRUE,
"enabled" = FALSE
WHERE id = $1
`

func (q *Queries) SoftDeleteUserRoleById(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, softDeleteUserRoleById, id)
	return err
}

const updateUserEmailById = `-- name: UpdateUserEmailById :one
UPDATE users
  set email = $2
WHERE id = $1
RETURNING id, username, password, email, created_at, last_modified, enabled, is_deleted
`

type UpdateUserEmailByIdParams struct {
	ID    int32
	Email pgtype.Text
}

func (q *Queries) UpdateUserEmailById(ctx context.Context, arg UpdateUserEmailByIdParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUserEmailById, arg.ID, arg.Email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Password,
		&i.Email,
		&i.CreatedAt,
		&i.LastModified,
		&i.Enabled,
		&i.IsDeleted,
	)
	return i, err
}

const updateUserPasswordById = `-- name: UpdateUserPasswordById :exec
UPDATE users
  set password = $2
WHERE id = $1
`

type UpdateUserPasswordByIdParams struct {
	ID       int32
	Password pgtype.Text
}

func (q *Queries) UpdateUserPasswordById(ctx context.Context, arg UpdateUserPasswordByIdParams) error {
	_, err := q.db.Exec(ctx, updateUserPasswordById, arg.ID, arg.Password)
	return err
}

const verifyUserPermissionById = `-- name: VerifyUserPermissionById :one
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
)
`

type VerifyUserPermissionByIdParams struct {
	UserId     pgtype.Int4
	Permission pgtype.Text
}

func (q *Queries) VerifyUserPermissionById(ctx context.Context, arg VerifyUserPermissionByIdParams) (bool, error) {
	row := q.db.QueryRow(ctx, verifyUserPermissionById, arg.UserId, arg.Permission)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const verifyUserPermissionByRoleId = `-- name: VerifyUserPermissionByRoleId :one
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
)
`

type VerifyUserPermissionByRoleIdParams struct {
	RoleId     int32
	Permission pgtype.Text
}

func (q *Queries) VerifyUserPermissionByRoleId(ctx context.Context, arg VerifyUserPermissionByRoleIdParams) (bool, error) {
	row := q.db.QueryRow(ctx, verifyUserPermissionByRoleId, arg.RoleId, arg.Permission)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}
