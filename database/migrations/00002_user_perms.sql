-- +goose Up
-- +goose StatementBegin
-- +goose envsub on
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $DEV_DB_USER;
SELECT 'sqlc parsing';
-- +goose envsub off
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- +goose envsub on
-- REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM $DEV_DB_USER;
-- +goose envsub off
-- +goose StatementEnd
