CREATE TABLE public.user_hosted_k8 (
    id integer NOT NULL,
    price_tier_code_id integer NOT NULL,
    user_id integer NOT NULL,
    organization_id integer,
    current_host_server_ids[] integer NOT NULL,
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
