-- +goose Up
-- +goose StatementBegin
-- +goose envsub on
-- GRANT pg_read_all_data TO $DEV_DB_USER;
-- GRANT pg_write_all_data TO $DEV_DB_USER;
SELECT "sqlc parsing";
-- +goose envsub on
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
