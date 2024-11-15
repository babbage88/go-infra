-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.user_hosted_db ALTER COLUMN db_platform_id SET DEFAULT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
