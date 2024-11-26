-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS users_idx_user_id ON users(id, username);
CREATE INDEX IF NOT EXISTS users_idx_role ON users("role");
CREATE INDEX IF NOT EXISTS users_idx_created ON users(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users_idx_user_id;
DROP INDEX IF EXISTS users_idx_role;
DROP INDEX IF EXISTS users_idx_created;
-- +goose StatementEnd
