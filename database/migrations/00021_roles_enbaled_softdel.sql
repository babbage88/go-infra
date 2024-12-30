-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.user_roles
ADD "enabled" bool DEFAULT true NOT NULL,
ADD "is_deleted" bool DEFAULT false NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.user_roles
DROP COLUMN "enabled",
DROP COLUMN "is_deleted";
-- +goose StatementEnd
