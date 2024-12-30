-- +goose Up
-- +goose StatementBegin
-- GRANT pg_read_all_data TO $DEV_DB_USER;
-- GRANT pg_write_all_data TO $DEV_DB_USER;
SELECT 'db parsing';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
