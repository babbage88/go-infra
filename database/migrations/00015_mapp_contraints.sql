-- +goose Up
-- +goose StatementBegin
SELECT 'buggy sqlc';
-- ALTER TABLE ONLY public.user_role_mapping
SELECT 'buggy sqlc';
--     ADD CONSTRAINT unique_user_role_id UNIQUE (user_id, role_id);
SELECT 'buggy sqlc';
-- ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- ALTER TABLE ONLY public.user_role_mapping
--     DROP CONSTRAINT unique_user_role_id;
-- +goose StatementEnd
