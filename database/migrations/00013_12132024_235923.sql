-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    devuser_id INTEGER;
BEGIN
    -- Find the user ID for the username "devuser"
    SELECT id INTO devuser_id FROM public.users WHERE username = 'devuser';


    -- Ensure the user exists
    IF devuser_id IS NOT NULL THEN
        -- Map the user to the Admin role (role_id = 999)
        INSERT INTO public.user_role_mapping (user_id, role_id)
        VALUES (devuser_id, 999)
        ON CONFLICT DO NOTHING;
    ELSE

        RAISE NOTICE 'No DEV_APP_USER found. Ensure correct env var is set. No mapping created.';

    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    -- Remove the user_role_mapping for the user "devuser" and Admin role (role_id = 999)
    DELETE FROM public.user_role_mapping

    WHERE user_id = (SELECT id FROM public.users WHERE username = 'devuser')
      AND role_id = 999;
END $$;
-- +goose StatementEnd
