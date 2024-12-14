-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.hosted_db_platforms (
    id integer NOT NULL,
    platform_name character varying(255) DEFAULT 'postgresql' NOT NULL,
    default_listen_port integer DEFAULT 5432 NULL

);

CREATE SEQUENCE IF NOT EXISTS public.hosted_db_platforms_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER SEQUENCE public.hosted_db_platforms_id_seq OWNED BY public.hosted_db_platforms.id;
ALTER TABLE ONLY public.hosted_db_platforms ALTER COLUMN id SET DEFAULT nextval('public.hosted_db_platforms_id_seq'::regclass);
ALTER TABLE ONLY public.hosted_db_platforms ADD CONSTRAINT hosted_db_platforms_pkey PRIMARY KEY (id);
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_type = 'FOREIGN KEY'
        AND table_name = 'user_hosted_db'
        AND constraint_name = 'user_hosted_db_platforms_id_fkey'
    ) THEN
        ALTER TABLE ONLY public.user_hosted_db
            ADD CONSTRAINT user_hosted_db_platforms_id_fkey FOREIGN KEY (db_platform_id) REFERENCES public.hosted_db_platforms(id);
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO public.hosted_db_platforms(id, platform_name, default_listen_port) VALUES
(1, 'postgresql', 5432),
(2, 'sqlite', NULL),
(3, 'mysql', 3306)
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE public.hosted_db_platforms;
-- +goose StatementEnd
