-- +goose Up
-- +goose StatementBegin
DROP VIEW IF EXISTS users_with_roles;
-- +goose StatementEnd
-- +goose StatementBegin
SELECT 'sqlc does not work with views apparently';
CREATE OR REPLACE  VIEW users_with_roles AS
SELECT
    u.id AS "id",
    u.username AS "username",
    u.password AS "password",
    u.email AS "email",
    COALESCE(ur.role_name, 'None') AS "role",
    COALESCE(urm.ROLE_ID, 0) AS "role_id",
    u.created_at AS "created_at",
    u.last_modified AS "last_modified",
    u.enabled AS "enabled",
    u.is_deleted AS "is_deleted"
FROM public.users u
LEFT JOIN public.user_role_mapping urm ON u.id = urm.user_id AND urm.enabled = TRUE
LEFT JOIN public.user_roles ur ON urm.role_id = ur.id
WHERE u.enabled = TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS users_with_roles;
-- +goose StatementEnd
