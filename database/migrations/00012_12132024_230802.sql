-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.user_roles (
    id integer NOT NULL,
    role_name character varying(255) NOT NULL,
    role_description text,
    expiration timestamp without time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT user_roles_pkey PRIMARY KEY (id),
    CONSTRAINT unique_role_name UNIQUE (role_name)

);

CREATE SEQUENCE IF NOT EXISTS public.user_roles_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.user_roles_id_seq OWNED BY public.user_roles.id;
ALTER TABLE ONLY public.user_roles ALTER COLUMN id SET DEFAULT nextval('public.user_roles_id_seq'::regclass);
CREATE INDEX user_roles_idx_created_at ON public.user_roles USING btree (created_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.app_permissions (
    id integer NOT NULL,
    permission_name character varying(255) NOT NULL,
    permission_description text,
    CONSTRAINT app_permissions_pkey PRIMARY KEY (id),
    CONSTRAINT unique_permission_name UNIQUE (permission_name)

);

CREATE SEQUENCE IF NOT EXISTS public.app_permissions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.app_permissions_id_seq OWNED BY public.app_permissions.id;
ALTER TABLE ONLY public.app_permissions ALTER COLUMN id SET DEFAULT nextval('public.app_permissions_id_seq'::regclass);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.user_role_mapping (
    id integer NOT NULL,
    user_id integer NOT NULL,
    role_id integer NOT NULL,
    "enabled" bool DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT user_role_mapping_pkey PRIMARY KEY (id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES public.users (id) ON DELETE CASCADE,
    CONSTRAINT fk_role FOREIGN KEY (role_id) REFERENCES public.user_roles (id) ON DELETE CASCADE,
    CONSTRAINT unique_user_role_id UNIQUE (user_id, role_id)
);

CREATE SEQUENCE IF NOT EXISTS public.user_role_mapping_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.user_role_mapping_id_seq OWNED BY public.user_role_mapping.id;
ALTER TABLE ONLY public.user_role_mapping ALTER COLUMN id SET DEFAULT nextval('public.user_role_mapping_id_seq'::regclass);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.role_permission_mapping (
    id integer NOT NULL,
    role_id integer NOT NULL,
    permission_id integer NOT NULL,
    enabled bool DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT role_permission_mapping_pkey PRIMARY KEY (id),
    CONSTRAINT fk_role FOREIGN KEY (role_id) REFERENCES public.user_roles (id) ON DELETE CASCADE,
    CONSTRAINT unique_perm_role_id UNIQUE (permission_id, role_id),
    CONSTRAINT fk_permission FOREIGN KEY (permission_id) REFERENCES public.app_permissions (id) ON DELETE CASCADE
);

CREATE SEQUENCE IF NOT EXISTS public.role_permission_mapping_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.role_permission_mapping_id_seq OWNED BY public.role_permission_mapping.id;
ALTER TABLE ONLY public.role_permission_mapping ALTER COLUMN id SET DEFAULT nextval('public.role_permission_mapping_id_seq'::regclass);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO public.user_roles (id, role_name, role_description) VALUES (999, 'Admin', 'Administrator role with all permissions')
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.app_permissions (permission_name, permission_description) VALUES
    ('CreateUser', 'Permission to create users'),
    ('AlterUser', 'Permission to alter users'),
    ('CreateDatabase', 'Permission to create databases')
ON CONFLICT (permission_name) DO NOTHING;

INSERT INTO public.role_permission_mapping (role_id, permission_id)
SELECT 999, id FROM public.app_permissions
WHERE permission_name IN ('CreateUser', 'AlterUser', 'CreateDatabase');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS public.user_roles ALTER COLUMN id DROP DEFAULT;
DROP TABLE IF EXISTS public.role_permission_mapping;
DROP SEQUENCE IF EXISTS public.role_permission_mapping_id_seq;
DROP TABLE IF EXISTS public.user_role_mapping;
DROP SEQUENCE IF EXISTS public.user_role_mapping_id_seq;
DROP TABLE IF EXISTS public.app_permissions;
DROP SEQUENCE IF EXISTS public.app_permissions_id_seq;
DROP TABLE IF EXISTS public.user_roles;
DROP SEQUENCE IF EXISTS public.user_roles_id_seq;
-- +goose StatementEnd

