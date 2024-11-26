-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD COLUMN "enabled" BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
DROP COLUMN IF EXISTS "enabled",
DROP COLUMN IF EXISTS is_deleted;
-- +goose StatementEnd
