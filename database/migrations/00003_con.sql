-- +goose Up
-- +goose StatementBegin
-- +goose envsub on
SELECT 'DB MIG';
-- ALTER ROLE $DEV_DB_USER WITH PASSWORD '${DEV_DB_USER_PW}';
-- +goose envsub off
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
