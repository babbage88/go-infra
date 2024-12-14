-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE VIEW user_permissions_view AS
SELECT
  u.id as "UserId",
  u.username as "Username",
  ap.id as "PermissionId",
  ap.permission_name as "Permission",
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
ORDER BY u.id Asc;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE VIEW user_roles_view AS
SELECT
  u.id as "UserId",
  u.username as "Username",
  ur.id as "RoleId",
  ur.role_name as "Role",
  urm.last_modified as "LastModified"
FROM
    public.user_role_mapping urm
LEFT JOIN
    user_roles ur on ur.id = urm.role_id
LEFT JOIN
    users u on u.id = urm.user_id
ORDER BY u.id Asc;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE VIEW role_permissions_view AS
SELECT
  ur.id as "RoleId",
  ur.role_name as "Role",
  ap.id AS "PermissionId",
  rpm.id AS "MappingId",
  ap.permission_name AS "Permission"
FROM
    user_roles ur
LEFT JOIN
    role_permission_mapping rpm on rpm.role_id = ur.id
LEFT JOIN
	app_permissions ap ON  rpm.role_id = ur.id
ORDER BY ap.id ;
-- +goose StatementEnd



-- +goose Down
-- +goose StatementBegin
-- ALTER TABLE public."user_role_mapping"
--     DROP COLUMN "enabled";
--
-- ALTER TABLE public."role_permission_mapping."
--     DROP COLUMN "enabled";
-- +goose StatementEnd