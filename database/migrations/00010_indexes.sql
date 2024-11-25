-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS auth_token_idx_userid ON auth_tokens(user_id);
CREATE INDEX IF NOT EXISTS auth_token_idx_created_at ON auth_tokens(created_at);
CREATE INDEX IF NOT EXISTS dns_records_idx_zone ON dns_records(zone_name);
CREATE INDEX IF NOT EXISTS dns_records_idx_name ON dns_records("name");
CREATE INDEX IF NOT EXISTS dns_records_idx_content ON dns_records(content);
CREATE INDEX IF NOT EXISTS dns_records_idx_type ON dns_records("type");
CREATE INDEX IF NOT EXISTS host_servers_idx_hostname ON host_servers(hostname);
CREATE INDEX IF NOT EXISTS host_servers_idx_username ON host_servers(username);
CREATE INDEX IF NOT EXISTS user_hosted_db_idx_userid ON user_hosted_db(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS auth_token_idx_userid;
DROP INDEX IF EXISTS auth_token_idx_created_at;
DROP INDEX IF EXISTS dns_records_idx_zone;
DROP INDEX IF EXISTS dns_records_idx_name;
DROP INDEX IF EXISTS dns_records_idx_content;
DROP INDEX IF EXISTS dns_records_idx_type;
DROP INDEX IF EXISTS host_servers_idx_hostname;
DROP INDEX IF EXISTS host_servers_idx_username;
DROP INDEX IF EXISTS user_hosted_db_idx_userid;
-- +goose StatementEnd
