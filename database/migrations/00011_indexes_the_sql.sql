-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS users_idx_enabled ON users("enabled");
CREATE INDEX IF NOT EXISTS users_idx_role ON users("role");
CREATE INDEX IF NOT EXISTS users_idx_created ON users(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users_idx_enabled;
DROP INDEX IF EXISTS users_idx_role;
DROP INDEX IF EXISTS users_idx_created;
-- +goose StatementEnd
