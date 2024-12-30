-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE VIEW public.users_with_roles AS
SELECT
    u.id,
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

-- +goose Down
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
     LEFT JOIN user_roles ur ON urm.role_id = ur.id;
-- +goose StatementEnd
