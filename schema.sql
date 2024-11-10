--
-- PostgreSQL database dump
--

-- Dumped from database version 15.8 (Debian 15.8-0+deb12u1)
-- Dumped by pg_dump version 15.8 (Debian 15.8-0+deb12u1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: auth_tokens; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.auth_tokens (
    id integer NOT NULL,
    user_id integer,
    token text,
    expiration timestamp without time zone,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.auth_tokens OWNER TO postgres;

--
-- Name: auth_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.auth_tokens_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.auth_tokens_id_seq OWNER TO postgres;

--
-- Name: auth_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.auth_tokens_id_seq OWNED BY public.auth_tokens.id;


--
-- Name: dns_records; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.dns_records (
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

--
-- Name: host_servers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.host_servers (
    id integer NOT NULL,
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

--
-- Name: host_servers_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.host_servers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.host_servers_id_seq OWNER TO postgres;

--
-- Name: host_servers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.host_servers_id_seq OWNED BY public.host_servers.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id integer NOT NULL,
    username character varying(255),
    password text,
    email character varying(255),
    role character varying(255),
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_modified timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: auth_tokens id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auth_tokens ALTER COLUMN id SET DEFAULT nextval('public.auth_tokens_id_seq'::regclass);


--
-- Name: host_servers id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.host_servers ALTER COLUMN id SET DEFAULT nextval('public.host_servers_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: auth_tokens auth_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auth_tokens
    ADD CONSTRAINT auth_tokens_pkey PRIMARY KEY (id);


--
-- Name: dns_records dns_records_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.dns_records
    ADD CONSTRAINT dns_records_pkey PRIMARY KEY (dns_record_id);


--
-- Name: host_servers host_servers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.host_servers
    ADD CONSTRAINT host_servers_pkey PRIMARY KEY (id);


--
-- Name: users unique_email; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT unique_email UNIQUE (email);


--
-- Name: host_servers unique_hostname_ip; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.host_servers
    ADD CONSTRAINT unique_hostname_ip UNIQUE (hostname, ip_address);


--
-- Name: users unique_username; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT unique_username UNIQUE (username);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: auth_tokens auth_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auth_tokens
    ADD CONSTRAINT auth_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);

--
-- Name: user_hosted_db; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_hosted_db (
    id integer NOT NULL,
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

--
-- Name: user_hosted_db_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_hosted_db_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_hosted_db_id_seq OWNER TO postgres;

--
-- Name: user_hosted_db_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_hosted_db_id_seq OWNED BY public.user_hosted_db.id;

--
-- Name: user_hosted_db id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_hosted_db ALTER COLUMN id SET DEFAULT nextval('public.user_hosted_db_id_seq'::regclass);

--
-- Name: user_hosted_db user_hosted_db_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_hosted_db
    ADD CONSTRAINT user_hosted_db_pkey PRIMARY KEY (id);

--
-- Name: user_hosted_db unique_hostname_ip; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_hosted_db
    ADD CONSTRAINT unique_pub_ip_port UNIQUE (pub_ip_address, listen_port);

CREATE TABLE public.user_hosted_k8 (
    id integer NOT NULL,
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

--
-- Name: user_hosted_k8_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_hosted_k8_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_hosted_k8_id_seq OWNER TO postgres;

--
-- Name: user_hosted_k8_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.user_hosted_k8_id_seq OWNED BY public.user_hosted_k8.id;

--
-- Name: user_hosted_k8 id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_hosted_k8 ALTER COLUMN id SET DEFAULT nextval('public.user_hosted_k8_id_seq'::regclass);

--
-- Name: user_hosted_k8 user_hosted_k8_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_hosted_k8
    ADD CONSTRAINT user_hosted_k8_pkey PRIMARY KEY (id);

--
-- Name: user_hosted_k8 unique_hostname_ip; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_hosted_k8
    ADD CONSTRAINT unique_pub_ip_port UNIQUE (pub_ip_address, listen_port);

