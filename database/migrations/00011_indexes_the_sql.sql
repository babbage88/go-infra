-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS users_idx_enabled ON users("enabled");
CREATE INDEX IF NOT EXISTS users_idx_role ON users("role");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users_idx_enabled;
DROP INDEX IF EXISTS users_idx_role;
-- +goose StatementEnd
