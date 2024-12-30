-- +goose Up
-- +goose StatementBegin
DROP VIEW public.users_with_roles;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW public.users_with_roles AS
SELECT
    u.id,
    u.username,
    u.password,
    u.email,
    COALESCE(array_agg(ur.role_name) FILTER (WHERE ur.role_name IS NOT NULL), ARRAY['None']) AS "roles",
    COALESCE(array_agg(urm.role_id) FILTER (WHERE urm.role_id IS NOT NULL), ARRAY[0]) AS role_ids,
    u.created_at,
    u.last_modified,
    u.enabled,
    u.is_deleted
FROM
    users u
LEFT JOIN
    user_role_mapping urm ON u.id = urm.user_id AND urm.enabled = true
LEFT JOIN
    user_roles ur ON urm.role_id = ur.id
WHERE
    u.is_deleted = false
GROUP BY
    u.id, u.username, u.password, u.email, u.created_at, u.last_modified, u.enabled, u.is_deleted;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW public.users_with_roles;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE VIEW public.users_with_roles
AS SELECT u.id,
    u.username,
    u.password,
    u.email,
    COALESCE(ur.role_name, 'None'::character varying) AS role,
    COALESCE(urm.role_id, 0) AS role_id,
    u.created_at,
    u.last_modified,
    u.enabled,
    u.is_deleted
   FROM users u
     LEFT JOIN user_role_mapping urm ON u.id = urm.user_id AND urm.enabled = true
     LEFT JOIN user_roles ur ON urm.role_id = ur.id
  WHERE u.is_deleted = false;
-- +goose StatementEnd
