-- +goose Up
-- +goose StatementBegin
DROP VIEW IF EXISTS user_roles_view;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW public.user_roles_active AS
SELECT id AS "RoleId",
    role_name AS "RoleName",
    role_description AS "RoleDescription",
    created_at AS "CreatedAt",
    last_modified AS "LastModified",
    "enabled" AS "Enabled",
    is_deleted AS "IsDeleted"
FROM public.user_roles
WHERE is_deleted IS FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS user_roles_active;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE VIEW public.user_roles_view
AS SELECT u.id AS "UserId",
    u.username AS "Username",
    ur.id AS "RoleId",
    ur.role_name AS "Role",
    urm.last_modified AS "LastModified"
   FROM user_role_mapping urm
     LEFT JOIN user_roles ur ON ur.id = urm.role_id
     LEFT JOIN users u ON u.id = urm.user_id
  ORDER BY u.id;
-- +goose StatementEnd
