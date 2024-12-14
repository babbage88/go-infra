-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE VIEW user_permissions_view AS
SELECT DISTINCT
  ap.id as "PermissionId",
  ap.permission_name as "Permission",
  u.username as "Username",
  u.id as "UserId",
  ur.role_name as "Role",
  urm.last_modified as "LastModified"
FROM
    public.user_role_mapping urm
LEFT JOIN
    user_roles ur on ur.id = urm.role_id
LEFT JOIN
    users u on u.id = urm.user_id
LEFT JOIN
    role_permission_mapping rpm on rpm.role_id = urm.role_id
LEFT JOIN
    app_permissions ap on ap.id = rpm.permission_id
ORDER BY u.id ASC;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE VIEW user_roles_view AS
SELECT DISTINCT
  u.id as "UserId",
  u.username as "Username",
  ur.id as "RoleId",
  ur.role_name as "RoleName",
  urm.last_modified as "LastModified"
FROM
    public.user_role_mapping urm
LEFT JOIN
    user_roles ur on ur.id = urm.role_id
LEFT JOIN
    users u on u.id = urm.user_id
ORDER BY u.id ASC;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE user_role_mapping
ADD COLUMN "enabled" BOOLEAN NOT NULL DEFAULT TRUE;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE role_permission_mapping
ADD COLUMN "enabled" BOOLEAN NOT NULL DEFAULT TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS user_permissions_view;
DROP VIEW IF EXISTS user_roles_view;

ALTER TABLE user_role_mapping
    DROP COLUMN "enabled";

ALTER TABLE role_permission_mapping
    DROP COLUMN "enabled";
-- +goose StatementEnd