-- +goose Up
-- +goose StatementBegin
DROP VIEW IF EXISTS role_permissions_view;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE VIEW role_permissions_view AS
SELECT ur.id AS "RoleId",
    ur.role_name AS "Role",
    ap.id AS "PermissionId",
    rpm.id AS "MappingId",
    ap.permission_name AS "Permission"
FROM user_roles ur
	LEFT JOIN role_permission_mapping rpm ON rpm.role_id = ur.id
    LEFT JOIN app_permissions ap ON rpm.role_id = ur.id
WHERE ur.enabled = true
ORDER BY ap.id;
-- +goose StatementEnd

-- +goose StatementBegin
DROP VIEW IF EXISTS user_permissions_view;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE VIEW user_permissions_view AS
SELECT
    u.id AS "UserId",
    u.username AS "Username",
    ap.id AS "PermissionId",
    ap.permission_name AS "Permission",
    ur.role_name AS "Role",
    urm.last_modified AS "LastModified"
FROM user_role_mapping urm
	LEFT JOIN user_roles ur ON ur.id = urm.role_id
    LEFT JOIN users u ON u.id = urm.user_id
    LEFT JOIN role_permission_mapping rpm ON rpm.role_id = urm.role_id
    LEFT JOIN app_permissions ap ON ap.id = rpm.permission_id
WHERE ur.enabled = TRUE
ORDER BY u.id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS role_permissions_view;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE VIEW role_permissions_view AS
SELECT ur.id AS "RoleId",
    ur.role_name AS "Role",
    ap.id AS "PermissionId",
    rpm.id AS "MappingId",
    ap.permission_name AS "Permission"
FROM user_roles ur
	LEFT JOIN role_permission_mapping rpm ON rpm.role_id = ur.id
    LEFT JOIN app_permissions ap ON rpm.role_id = ur.id
ORDER BY ap.id;
-- +goose StatementEnd

-- +goose StatementBegin
DROP VIEW IF EXISTS user_permissions_view;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE VIEW user_permissions_view AS
SELECT
    u.id AS "UserId",
    u.username AS "Username",
    ap.id AS "PermissionId",
    ap.permission_name AS "Permission",
    ur.role_name AS "Role",
    urm.last_modified AS "LastModified"
FROM user_role_mapping urm
	LEFT JOIN user_roles ur ON ur.id = urm.role_id
    LEFT JOIN users u ON u.id = urm.user_id
    LEFT JOIN role_permission_mapping rpm ON rpm.role_id = urm.role_id
    LEFT JOIN app_permissions ap ON ap.id = rpm.permission_id
ORDER BY u.id;
-- +goose StatementEnd