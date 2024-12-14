-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE VIEW users_view AS
SELECT
    u."id" as "id",
    u.username as "username",
    u."password" as "password",
    u.email as "email",
    urv."RoleName" as "role",
    u.created_at as "created_at",
    u.LAST_MODIFIED,
    u."enabled" as "enabled",
    u.IS_DELETED as "is_deleted"
FROM public.users u
LEFT JOIN USER_ROLES_VIEW URV on "UserId" = u.ID
ORDER BY u.id ASC;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS users_view;
-- +goose StatementEnd
