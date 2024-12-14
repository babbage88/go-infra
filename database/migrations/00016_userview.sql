-- +goose Up
-- +goose StatementBegin
SELECT 'sqlc does not work with views apparently';
CREATE OR REPLACE  VIEW users_with_roles AS
SELECT
    u.id AS "id",
    u.username AS "username",
    u.password AS "password",
    u.email AS "email",
    ur.role_name AS "role", -- Join the role_name from user_roles
    u.created_at AS "created_at",
    u.last_modified AS "last_modified",
    u.enabled AS "enabled",
    u.is_deleted AS "is_deleted"
FROM public.users u
LEFT JOIN public.user_role_mapping urm ON u.id = urm.user_id
LEFT JOIN public.user_roles ur ON urm.role_id = ur.id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS users_with_roles;
-- +goose StatementEnd
