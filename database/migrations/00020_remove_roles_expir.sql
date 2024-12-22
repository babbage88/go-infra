-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.user_roles
DROP COLUMN "expiration";
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.user_roles
ADD "expiration" timestamp NULL;
-- +goose StatementEnd
