-- +goose no transaction
-- +goose envsub on
CREATE DATABASE ${DEV_DATABASE_NAME}
-- +goose envsub off

-- +goose Up
-- +goose StatementBegin
-- +goose envsub on

CREATE ROLE ${DEV_DB_USER} WITH
	CREATEROLE
	INHERIT
	LOGIN
	NOBYPASSRLS
	CONNECTION LIMIT -1;

GRANT UPDATE, DELETE, INSERT, SELECT ON TABLE public.auth_tokens TO ${DEV_DB_USER} WITH GRANT OPTION;
GRANT UPDATE, DELETE, INSERT, SELECT ON TABLE public.dns_records TO ${DEV_DB_USER} WITH GRANT OPTION;
GRANT UPDATE, DELETE, INSERT, SELECT ON TABLE public.host_servers TO ${DEV_DB_USER} WITH GRANT OPTION;
GRANT UPDATE, DELETE, INSERT, SELECT ON TABLE public.users TO ${DEV_DB_USER} WITH GRANT OPTION;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE public.auth_tokens (
    id integer not NULL,
    user_id integer,
    token text,
    expiration timestamp without time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE public.auth_tokens OWNER TO postgres;

CREATE SEQUENCE public.auth_tokens_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE public.auth_tokens_id_seq OWNER TO postgres;
ALTER SEQUENCE public.auth_tokens_id_seq OWNED BY public.auth_tokens.id;

CREATE TABLE public.dns_records (
    id integer not NULL,
    dns_record_id text NOT NULL,
    zone_name text,
    zone_id text,
    name text,
    content text,
    proxied boolean,
    type character varying(10),
    comment text,
    ttl integer,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE public.dns_records OWNER TO postgres;

CREATE SEQUENCE public.dns_records_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE public.dns_records_id_seq OWNER TO postgres;
ALTER SEQUENCE public.dns_records_id_seq OWNED BY public.dns_records.id;

CREATE TABLE public.host_servers (
    id integer not NULL,
    hostname character varying(255) NOT NULL,
    ip_address inet NOT NULL,
    username character varying(255),
    public_ssh_keyname character varying(255),
    hosted_domains character varying(255)[],
    ssl_key_path character varying(4098),
    is_container_host boolean,
    is_vm_host boolean,
    is_virtual_machine boolean,
    id_db_host boolean,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE public.host_servers OWNER TO postgres;

CREATE SEQUENCE public.host_servers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE public.host_servers_id_seq OWNER TO postgres;
ALTER SEQUENCE public.host_servers_id_seq OWNED BY public.host_servers.id;

CREATE TABLE public.users (
    id integer not NULL,
    username character varying(255),
    password text,
    email character varying(255),
    role character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE public.users OWNER TO postgres;

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE public.users_id_seq OWNER TO postgres;
ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

CREATE TABLE public.user_hosted_k8  (
    id integer not NULL,
    price_tier_code_id integer NOT NULL,
    user_id integer NOT NULL,
    organization_id integer,
    current_host_server_ids integer[] NOT NULL,
    user_application_ids integer[],
    user_certificate_ids integer[],
    k8_type character varying(255) DEFAULT 'k3s' NOT NULL,
    api_endpoint_fqdn character varying(255) NOT NULL,
    cluster_name character varying(255) NOT NULL,
    pub_ip_address inet NOT NULL,
    listen_port integer NOT NULL,
    private_ip_address inet,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE public.user_hosted_k8 OWNER TO postgres;

CREATE SEQUENCE public.user_hosted_k8_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE public.user_hosted_k8_id_seq OWNER TO postgres;
ALTER SEQUENCE public.user_hosted_k8_id_seq OWNED BY public.user_hosted_k8.id;

CREATE TABLE public.user_hosted_db (
    id integer not NULL,
    price_tier_code_id integer NOT NULL,
    user_id integer NOT NULL,
    current_host_server_id integer NOT NULL,
    current_kube_cluster_id integer,
    user_application_ids integer[],
    db_platform_id integer NOT NULL,
    fqdn character varying(255) NOT NULL,
    pub_ip_address inet NOT NULL,
    listen_port integer NOT NULL,
    private_ip_address inet,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

ALTER TABLE public.user_hosted_db OWNER TO postgres;

CREATE SEQUENCE public.user_hosted_db_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE public.user_hosted_db_id_seq OWNER TO postgres;

ALTER SEQUENCE public.user_hosted_db_id_seq OWNED BY public.user_hosted_db.id;

ALTER TABLE ONLY public.user_hosted_db
    ADD CONSTRAINT unique_pub_ip_port UNIQUE (pub_ip_address, listen_port);

ALTER TABLE ONLY public.auth_tokens ALTER COLUMN id SET DEFAULT nextval('public.auth_tokens_id_seq'::regclass);
ALTER TABLE ONLY public.dns_records ALTER COLUMN id SET DEFAULT nextval('public.dns_records_id_seq'::regclass);
ALTER TABLE ONLY public.host_servers ALTER COLUMN id SET DEFAULT nextval('public.host_servers_id_seq'::regclass);
ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);
ALTER TABLE ONLY public.user_hosted_db ALTER COLUMN id SET DEFAULT nextval('public.user_hosted_db_id_seq'::regclass);
ALTER TABLE ONLY public.user_hosted_k8 ALTER COLUMN id SET DEFAULT nextval('public.user_hosted_k8_id_seq'::regclass);

ALTER TABLE ONLY public.auth_tokens
    ADD CONSTRAINT auth_tokens_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.dns_records
    ADD CONSTRAINT dns_records_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.host_servers
    ADD CONSTRAINT host_servers_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_hosted_db
    ADD CONSTRAINT user_hosted_db_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.user_hosted_k8
    ADD CONSTRAINT user_hosted_k8_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT unique_email UNIQUE (email);

ALTER TABLE ONLY public.host_servers
    ADD CONSTRAINT unique_hostname_ip UNIQUE (hostname, ip_address);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT unique_username UNIQUE (username);

ALTER TABLE ONLY public.auth_tokens
    ADD CONSTRAINT auth_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);
-- +goose StatementEnd









-- +goose envsub off
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
