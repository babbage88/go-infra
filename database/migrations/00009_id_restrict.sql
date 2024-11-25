-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ADD CONSTRAINT check_usesr_id_nonzero CHECK (id > 0);

ALTER TABLE auth_tokens
ADD CONSTRAINT check_auth_token_id_nonzero CHECK (id > 0);

ALTER TABLE dns_records
ADD CONSTRAINT check_dns_id_nonzero CHECK (id > 0);

ALTER TABLE host_servers
ADD CONSTRAINT check_servers_id_nonzero CHECK (id > 0);

ALTER TABLE user_hosted_db
ADD CONSTRAINT check_hosted_db_id_nonzero CHECK (id > 0);

ALTER TABLE user_hosted_k8
ADD CONSTRAINT check_hosted_k8_id_nonzero CHECK (id > 0);

ALTER TABLE hosted_db_platforms
ADD CONSTRAINT check_db_platforms_id_nonzero CHECK (id > 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
DROP CONSTRAINT IF EXISTS check_usesr_id_nonzero;

ALTER TABLE auth_tokens
DROP CONSTRAINT IF EXISTS check_auth_token_id_nonzero;

ALTER TABLE dns_records
DROP CONSTRAINT IF EXISTS check_dns_id_nonzero;

ALTER TABLE host_servers
DROP CONSTRAINT IF EXISTS check_servers_id_nonzero;

ALTER TABLE user_hosted_db
DROP CONSTRAINT IF EXISTS check_hosted_db_id_nonzero;

ALTER TABLE user_hosted_k8
DROP CONSTRAINT IF EXISTS check_hosted_k8_id_nonzero;

ALTER TABLE hosted_db_platforms
DROP CONSTRAINT IF EXISTS check_db_platforms_id_nonzero;
-- +goose StatementEnd

