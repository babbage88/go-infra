-- +goose Up
-- +goose StatementBegin
CREATE TABLE users_audit (
    audit_id SERIAL PRIMARY KEY,
    user_id INT,
    username character varying(255),
    email character varying(255),
    deleted_at TIMESTAMPTZ DEFAULT now(),
    deleted_by character varying(255)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION log_user_deletion()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO users_audit (user_id, username, email, deleted_at, deleted_by)
    VALUES (OLD.id, OLD.username, OLD.email, now(), current_user);
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER user_delete_trigger
AFTER DELETE ON users
FOR EACH ROW
EXECUTE FUNCTION log_user_deletion();
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TRIGGER user_delete_trigger ON public.users;
DROP FUNCTION log_user_deletion;
DROP TABLE users_audit;
-- +goose StatementEnd
