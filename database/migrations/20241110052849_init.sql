-- +goose no transaction
-- +goose up
-- +goose envsub on
CREATE DATABASE $DEV_DATABASE_NAME;
-- +goose envsub off

-- +goose down
-- +goose envsub on
DROP DATABASE $DEV_DATABASE_NAME;
-- +goose envsub off

