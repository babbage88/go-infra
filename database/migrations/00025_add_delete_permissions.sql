-- +goose Up
-- +goose StatementBegin
INSERT INTO public.app_permissions (permission_name, permission_description) VALUES
    ('DelteUser', 'Delete Users'),
    ('DeleteRole', 'Delete Roles'),
    ('DeltePermission', 'Delete Permissions'),
    ('DeleteDatabase', 'Delete Hosted Database')
ON CONFLICT (permission_name) DO NOTHING;

INSERT INTO public.role_permission_mapping (role_id, permission_id)
SELECT 999, id FROM public.app_permissions
WHERE permission_name IN ('DeleteUser', 'DeleteRole', 'DeletePermission', 'DeleteDatabase');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
