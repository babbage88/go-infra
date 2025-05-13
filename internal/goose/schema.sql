-- public.users definition

CREATE TABLE public.users (
	id serial4 NOT NULL,
	username varchar(255) NULL,
	"password" text NULL,
	email varchar(255) NULL,
	created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
	last_modified timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
	enabled bool DEFAULT true NOT NULL,
	is_deleted bool DEFAULT false NOT NULL,
	CONSTRAINT check_usesr_id_nonzero CHECK ((id > 0)),
	CONSTRAINT unique_email UNIQUE (email),
	CONSTRAINT unique_username UNIQUE (username),
	CONSTRAINT users_pkey PRIMARY KEY (id)
);
CREATE INDEX users_idx_created ON public.users USING btree (created_at);
CREATE INDEX users_idx_user_id ON public.users USING btree (id, username);

-- Table Triggers

create trigger user_delete_trigger after
delete
    on
    public.users for each row execute function log_user_deletion();

-- public.app_permissions definition

CREATE TABLE public.app_permissions (
	id serial4 NOT NULL,
	permission_name varchar(255) NOT NULL,
	permission_description text NULL,
	CONSTRAINT app_permissions_pkey PRIMARY KEY (id),
	CONSTRAINT unique_permission_name UNIQUE (permission_name)
); 

-- public.role_permission_mapping definition

CREATE TABLE public.role_permission_mapping (
	id serial4 NOT NULL,
	role_id int4 NOT NULL,
	permission_id int4 NOT NULL,
	enabled bool DEFAULT true NOT NULL,
	created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
	last_modified timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
	CONSTRAINT role_permission_mapping_pkey PRIMARY KEY (id),
	CONSTRAINT unique_perm_role_id UNIQUE (permission_id, role_id)
);


-- public.role_permission_mapping foreign keys

ALTER TABLE public.role_permission_mapping ADD CONSTRAINT fk_permission FOREIGN KEY (permission_id) REFERENCES public.app_permissions(id) ON DELETE CASCADE;
ALTER TABLE public.role_permission_mapping ADD CONSTRAINT fk_role FOREIGN KEY (role_id) REFERENCES public.user_roles(id) ON DELETE CASCADE;