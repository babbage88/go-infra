-- +goose Up
-- +goose StatementBegin
INSERT INTO public.app_permissions (permission_name, permission_description) VALUES
    ('CreateRole', 'Permission to Create User Roles'),
    ('AlterRole', 'Alter Role properties and Permissions to Create User Roles'),
    ('CreatePermission', 'Create Permission'),
    ('AlterPermission', 'Alter Permission properties')
ON CONFLICT (permission_name) DO NOTHING;

INSERT INTO public.role_permission_mapping (role_id, permission_id)
SELECT 999, id FROM public.app_permissions
WHERE permission_name IN ('CreateRole', 'AlterRole', 'CreatePermission', 'AlterPermission');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
