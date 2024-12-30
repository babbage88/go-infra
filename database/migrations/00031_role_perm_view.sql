-- +goose Up
-- +goose StatementBegin
DROP VIEW public.role_permissions_view;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE VIEW public.role_permissions_view
AS SELECT
    ur.id AS "RoleId",
    ur.role_name AS "Role",
    ap.id AS "PermissionId",
    ap.permission_name AS "Permission"
FROM
    user_roles ur
LEFT JOIN
    role_permission_mapping rpm
    ON rpm.role_id = ur.id
LEFT JOIN
    app_permissions ap
    ON rpm.permission_id = ap.id
WHERE
    ur.enabled = true
    AND rpm.enabled = true
GROUP BY
    ur.id, ur.role_name, ap.id, ap.permission_name
ORDER BY
    ur.id, ap.id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW public.role_permissions_view;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE VIEW public.role_permissions_view
AS SELECT ur.id AS "RoleId",
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
