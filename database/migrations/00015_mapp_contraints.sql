-- +goose Up
-- +goose StatementBegin
ALTER TABLE ONLY public.user_role_mapping
    ADD CONSTRAINT unique_user_role_id UNIQUE (user_id, role_id);
;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ONLY public.user_role_mapping
    DROP CONSTRAINT unique_user_role_id;
-- +goose StatementEnd
