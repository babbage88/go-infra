-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    devuser_id INTEGER;
BEGIN
    -- Find the user ID for the username "devuser"
-- +goose envsub on
    SELECT id INTO devuser_id FROM public.users WHERE username = '$DEV_APP_USER';
-- +goose envsub off

    -- Ensure the user exists
    IF devuser_id IS NOT NULL THEN
        -- Map the user to the Admin role (role_id = 999)
        INSERT INTO public.user_role_mapping (user_id, role_id)
        VALUES (devuser_id, 999)
        ON CONFLICT DO NOTHING;
    ELSE
-- +goose envsub on
        RAISE NOTICE 'User "$DEV_APP_USER" not found. No mapping created.';
-- +goose envsub off
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    -- Remove the user_role_mapping for the user "devuser" and Admin role (role_id = 999)
    DELETE FROM public.user_role_mapping
-- +goose envsub on
    WHERE user_id = (SELECT id FROM public.users WHERE username = '$DEV_APP_USER')
-- +goose envsub off
      AND role_id = 999;
END $$;
-- +goose StatementEnd
