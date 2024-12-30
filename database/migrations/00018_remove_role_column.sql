-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.users
DROP COLUMN "role";
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.users
ADD "role" varchar(255) NULL;
-- +goose StatementEnd
