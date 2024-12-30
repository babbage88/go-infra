-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.auth_tokens (
    id integer not NULL,
    user_id integer,
    token text,
    expiration timestamp without time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE SEQUENCE IF NOT EXISTS public.auth_tokens_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.auth_tokens_id_seq OWNED BY public.auth_tokens.id;

CREATE TABLE IF NOT EXISTS public.dns_records (
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

CREATE SEQUENCE IF NOT EXISTS public.dns_records_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.dns_records_id_seq OWNED BY public.dns_records.id;

CREATE TABLE IF NOT EXISTS public.host_servers (
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

CREATE SEQUENCE IF NOT EXISTS public.host_servers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.host_servers_id_seq OWNED BY public.host_servers.id;

CREATE TABLE IF NOT EXISTS public.users (
    id integer not NULL,
    username character varying(255),
    password text,
    email character varying(255),
    role character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE SEQUENCE IF NOT EXISTS public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

CREATE TABLE IF NOT EXISTS public.user_hosted_k8  (
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

CREATE SEQUENCE IF NOT EXISTS public.user_hosted_k8_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.user_hosted_k8_id_seq OWNED BY public.user_hosted_k8.id;

CREATE TABLE IF NOT EXISTS public.user_hosted_db (
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

CREATE SEQUENCE IF NOT EXISTS public.user_hosted_db_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE public.user_hosted_db_id_seq OWNED BY public.user_hosted_db.id;

ALTER TABLE ONLY public.user_hosted_db
    ADD CONSTRAINT unique_pub_ip_port UNIQUE (pub_ip_address, listen_port);

ALTER TABLE ONLY public.auth_tokens ALTER COLUMN id SET DEFAULT nextval('public.auth_tokens_id_seq'::regclass);
ALTER TABLE ONLY public.dns_records ALTER COLUMN id SET DEFAULT nextval('public.dns_records_id_seq'::regclass);
ALTER TABLE ONLY public.host_servers ALTER COLUMN id SET DEFAULT nextval('public.host_servers_id_seq'::regclass);
ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);
ALTER TABLE ONLY public.user_hosted_db ALTER COLUMN id SET DEFAULT nextval('public.user_hosted_db_id_seq'::regclass);
ALTER TABLE ONLY public.user_hosted_k8 ALTER COLUMN id SET DEFAULT nextval('public.user_hosted_k8_id_seq'::regclass);
ALTER TABLE public.user_hosted_db OWNER TO postgres;
ALTER TABLE public.auth_tokens OWNER TO postgres;
ALTER TABLE public.dns_records OWNER TO postgres;
ALTER TABLE public.user_hosted_k8 OWNER TO postgres;
ALTER TABLE public.host_servers OWNER TO postgres;
ALTER TABLE public.users OWNER TO postgres;

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

-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_type = 'FOREIGN KEY'
        AND table_name = 'auth_tokens'
        AND constraint_name = 'auth_tokens_user_id_fkey'
    ) THEN
        ALTER TABLE ONLY public.auth_tokens
            ADD CONSTRAINT auth_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);
    END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin

-- DO $$
-- BEGIN
--     IF NOT EXISTS (
--         SELECT 1 FROM pg_roles
-- -- +goose envsub on
--         WHERE rolname = '${DEV_DB_USER}'
--     ) THEN
--         CREATE ROLE ${DEV_DB_USER};
-- -- +goose envsub off
--     END IF;
-- END $$;
-- +goose StatementEnd

-- +goose StatementBegin
-- +goose envsub on
-- Creating go-infra app/api user for development/testing
 INSERT INTO public.users
 (id, username, "password", email, "role", created_at, last_modified)
 VALUES(nextval('users_id_seq'::regclass), 'devuser', '${DEV_APP_USER_PW}', 'devuser@test.trahan.dev', 'admin', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

INSERT INTO public.dns_records (dns_record_id,zone_name,zone_id,"name","content",proxied,"type","comment",ttl,created_at,last_modified) VALUES
	 ('c7b13dc759c2be9ba60f3e4f2af764ff','balloonstx.com','dd9ea021cc2896395ab45c975389f1aa','autodiscover.balloonstx.com','autodiscover.outlook.com',true,'CNAME','',1,'2024-05-25 14:05:54.48794-05','2024-05-25 14:05:54.48794-05'),
	 ('a1686cb6272617afe79b9457174a6c67','balloonstx.com','dd9ea021cc2896395ab45c975389f1aa','balloonstx.com','balloonstx-com.mail.protection.outlook.com',false,'MX','',1,'2024-05-25 14:05:54.48794-05','2024-05-25 14:05:54.48794-05'),
	 ('7cf0992ab01e68e3197a54a77b7e09ef','balloonstx.com','dd9ea021cc2896395ab45c975389f1aa','_acme-challenge.balloonstx.com','AZw43t1F1JAFiBsbCLvIJMIZWCuBy-yphmh4HfIkYjY',false,'TXT','',1,'2024-05-25 14:05:54.48794-05','2024-05-25 14:05:54.48794-05'),
	 ('65bf522bf9f254cf6a1746971733bb2d','balloonstx.com','dd9ea021cc2896395ab45c975389f1aa','balloonstx.com','"v=spf1 include:spf.protection.outlook.com -all"',false,'TXT','',1,'2024-05-25 14:05:54.48794-05','2024-05-25 14:05:54.48794-05'),
	 ('e4410c35b4d4dedec6c815e2eb167124','balloonstx.com','dd9ea021cc2896395ab45c975389f1aa','balloonstx.com','"MS=ms68911349"',false,'TXT','',1,'2024-05-25 14:05:54.48794-05','2024-05-25 14:05:54.48794-05'),
	 ('943b050aaba009ad580e8fbf013d624b','trahan.dev','1a03a1886dc5855341b01d0afa9fa3c3','autodiscover.trahan.dev','autodiscover.outlook.com',true,'CNAME','',1,'2024-05-25 14:05:54.48794-05','2024-06-23 16:38:17.466044-05'),
	 ('ec819fbfb98e263b09ce3517786b3445','trahan.dev','1a03a1886dc5855341b01d0afa9fa3c3','calc.trahan.dev','react-trahan-compound.pages.dev',true,'CNAME','',1,'2024-05-25 14:05:54.48794-05','2024-06-23 16:38:17.469496-05'),
	 ('3f5d7758ae6a5ca2047fe233aca52abf','trahan.dev','1a03a1886dc5855341b01d0afa9fa3c3','compound.trahan.dev','react-trahan-compound.pages.dev',true,'CNAME','',1,'2024-05-25 14:05:54.48794-05','2024-06-23 16:38:17.47152-05'),
	 ('a247bf2a426ba35a365839e626e9554a','trahan.dev','1a03a1886dc5855341b01d0afa9fa3c3','trahan.dev','trahan-dev.mail.protection.outlook.com',false,'MX','',1,'2024-05-25 14:05:54.48794-05','2024-06-23 16:38:17.473594-05'),
	 ('95ce32b71078d5b4552648045752c273','trahan.dev','1a03a1886dc5855341b01d0afa9fa3c3','trahan.dev','"N3FoVWdlPvhYxICjyzb7kGpeZJyipMAsPBuweVJ0P3M"',false,'TXT','',1,'2024-05-25 14:05:54.48794-05','2024-06-23 16:38:17.475724-05');

INSERT INTO public.dns_records (dns_record_id,zone_name,zone_id,"name","content",proxied,"type","comment",ttl,created_at,last_modified) VALUES
	 ('b5c78d789c6ce3edde0443ea0342d380','trahan.dev','1a03a1886dc5855341b01d0afa9fa3c3','trahan.dev','"v=spf1 include:spf.protection.outlook.com -all"',false,'TXT','',1,'2024-05-25 14:05:54.48794-05','2024-06-23 16:38:17.477628-05'),
	 ('0e7113d57651515ec487d4170b2992b8','justintrahan.com','8b27cf3f30abe6dbfd7c10d94db5509b','api.justintrahan.com','104.8.87.242',true,'A','',1,'2024-05-25 14:05:54.48794-05','2024-05-25 14:05:54.48794-05'),
	 ('a9b9dc36a5b2f21e776eb3fa9df5ca7f','justintrahan.com','8b27cf3f30abe6dbfd7c10d94db5509b','justintrahan.com','104.8.87.242',true,'A','',1,'2024-05-25 14:05:54.48794-05','2024-05-25 14:05:54.48794-05'),
	 ('8fc3eff7e29110d7f0217a7a93620a5a','justintrahan.com','8b27cf3f30abe6dbfd7c10d94db5509b','plex.justintrahan.com','104.8.87.242',true,'A','',1,'2024-05-25 14:05:54.48794-05','2024-05-25 14:05:54.48794-05'),
	 ('8cef3c91029049868cde1d1a53f84d13','justintrahan.com','8b27cf3f30abe6dbfd7c10d94db5509b','www.justintrahan.com','104.8.87.242',true,'A','',1,'2024-05-25 14:05:54.48794-05','2024-05-25 14:05:54.48794-05'),
	 ('3a6e92a22d08c20fe69eb51f9a41a9bb','justintrahan.com','8b27cf3f30abe6dbfd7c10d94db5509b','autodiscover.justintrahan.com','autodiscover.outlook.com',true,'CNAME','',1,'2024-05-25 14:05:54.48794-05','2024-05-25 14:05:54.48794-05'),
	 ('2f565a662fd5e6f1cf0386a18991152d','trahan.dev','1a03a1886dc5855341b01d0afa9fa3c3','trahan.dev','"MS=ms42007862"',false,'TXT','',1,'2024-05-25 14:05:54.48794-05','2024-06-23 16:38:17.479621-05'),
	 ('8e84ba51ce98f90f181e20a2ffef3982','justintrahan.com','8b27cf3f30abe6dbfd7c10d94db5509b','justintrahan.com','justintrahan-com.mail.protection.outlook.com',false,'MX','',1,'2024-05-25 14:05:54.48794-05','2024-05-25 14:05:54.48794-05'),
	 ('6f45d4b1798c1a5b786ab5676d17dfa0','justintrahan.com','8b27cf3f30abe6dbfd7c10d94db5509b','_acme-challenge.justintrahan.com','graSlIjsnBXLdsY-mWLXxr14t0aSA7HOCBzw4InAVA0',false,'TXT','',60,'2024-05-25 14:05:54.48794-05','2024-05-25 14:05:54.48794-05'),
	 ('3b19e96430fe4effb0a4cd3d28285a72','justintrahan.com','8b27cf3f30abe6dbfd7c10d94db5509b','justintrahan.com','"v=spf1 include:spf.protection.outlook.com -all"',false,'TXT','',1,'2024-05-25 14:05:54.48794-05','2024-05-25 14:05:54.48794-05');

INSERT INTO public.dns_records (dns_record_id,zone_name,zone_id,"name","content",proxied,"type","comment",ttl,created_at,last_modified) VALUES
	 ('935723d24a543b251072917977c77558','justintrahan.com','8b27cf3f30abe6dbfd7c10d94db5509b','justintrahan.com','"MS=ms38949339"',false,'TXT','',1,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

INSERT INTO public.host_servers (hostname,ip_address,username,public_ssh_keyname,hosted_domains,ssl_key_path,is_container_host,is_vm_host,is_virtual_machine,id_db_host,created_at,last_modified) VALUES
	 ('trah-dev-01.trahan.dev','10.0.0.32'::inet,'jtrahan','/home/jtrahan/.ssh/id_rsa','{git.trahan.dev,calc.test.trahan.dev}','/etc/letsencrypt/live',true,false,true,false,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
	 ('trahvps1.trahan.dev','10.0.0.32'::inet,'jtrahan','/home/jtrahan/.ssh/id_rsa','{calc.test.trahan.dev}','',true,false,true,false,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
-- +goose envsub off
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.auth_tokens;
DROP TABLE IF EXISTS public.dns_records;
DROP TABLE IF EXISTS public.host_servers;
DROP TABLE IF EXISTS public.users;
DROP TABLE IF EXISTS public.user_hosted_db;
DROP TABLE IF EXISTS public.user_hosted_k8;
-- +goose StatementEnd
