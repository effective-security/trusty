BEGIN;

create or replace function create_constraint_if_not_exists (
    s_name text, t_name text, c_name text, constraint_sql text
)
returns void AS
$$
begin
    -- Look for our constraint
    if not exists (select constraint_name
                   from information_schema.constraint_column_usage
                   where table_schema = s_name and table_name = t_name and constraint_name = c_name) then
        execute constraint_sql;
    end if;
end;
$$ language 'plpgsql';

--
-- USERS
--
CREATE TABLE IF NOT EXISTS public.users
(
    id bigint NOT NULL,
    extern_id character varying(32) COLLATE pg_catalog."default" NOT NULL,
    provider character varying(16) COLLATE pg_catalog."default" NOT NULL,
    login character varying(64) COLLATE pg_catalog."default" NOT NULL,
    name character varying(64) COLLATE pg_catalog."default" NOT NULL,
    email character varying(160) COLLATE pg_catalog."default" NOT NULL,
    company character varying(64) COLLATE pg_catalog."default" NULL,
    avatar_url character varying(256) COLLATE pg_catalog."default" NULL,
    access_token text COLLATE pg_catalog."default" NULL,
    refresh_token text COLLATE pg_catalog."default" NULL,
    token_expires_at timestamp with time zone,
    login_count integer,
    last_login_at timestamp with time zone,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_provider_extern_id UNIQUE (provider, extern_id),
    CONSTRAINT users_provider_login UNIQUE (provider, login),
    CONSTRAINT users_provider_email UNIQUE (provider, email)
)
WITH (
    OIDS = FALSE
);

ALTER TABLE public.users
    OWNER to postgres;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_provider_email
    ON public.users USING btree
    (provider,email COLLATE pg_catalog."default")
    ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_provider_login
    ON public.users USING btree
    (provider,login COLLATE pg_catalog."default")
    ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_extern_id
    ON public.users USING btree
    (provider,extern_id)
    ;

SELECT create_constraint_if_not_exists(
    'public',
    'users',
    'unique_users_provider_login',
    'ALTER TABLE public.users ADD CONSTRAINT unique_users_provider_login UNIQUE USING INDEX idx_users_provider_login;');

SELECT create_constraint_if_not_exists(
    'public',
    'users',
    'unique_users_provider_email',
    'ALTER TABLE public.users ADD CONSTRAINT unique_users_provider_email UNIQUE USING INDEX idx_users_provider_email;');

--
-- ORGANIZATIONS
--
CREATE TABLE IF NOT EXISTS public.orgs
(
    id bigint NOT NULL,
    extern_id character varying(32) COLLATE pg_catalog."default" NOT NULL,
    provider character varying(16) COLLATE pg_catalog."default" NOT NULL,
    login character varying(64) COLLATE pg_catalog."default" NOT NULL,
    name character varying(64) COLLATE pg_catalog."default" NOT NULL,
    email character varying(160) COLLATE pg_catalog."default" NOT NULL,
    billing_email character varying(160) COLLATE pg_catalog."default" NOT NULL,
    company character varying(64) COLLATE pg_catalog."default" NULL,
    location character varying(64) COLLATE pg_catalog."default" NULL,
    avatar_url character varying(256) COLLATE pg_catalog."default" NULL,
    html_url character varying(256) COLLATE pg_catalog."default" NULL,
    type character varying(16) COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    street_address character varying(256) COLLATE pg_catalog."default" NULL,
    city character varying(32) COLLATE pg_catalog."default" NULL,
    postal_code character varying(16) COLLATE pg_catalog."default" NULL,
    region character varying(16) COLLATE pg_catalog."default" NULL,
    country character varying(16) COLLATE pg_catalog."default" NULL,
    phone character varying(32) COLLATE pg_catalog."default" NULL,
    CONSTRAINT orgs_pkey PRIMARY KEY (id),
    CONSTRAINT orgs_provider_extern_id UNIQUE (provider, extern_id),
    CONSTRAINT orgs_provider_login UNIQUE (provider, login)
)
WITH (
    OIDS = FALSE
);

ALTER TABLE public.orgs
    OWNER to postgres;

CREATE INDEX IF NOT EXISTS idx_orgs_provider
    ON public.orgs USING btree
    (provider);

CREATE INDEX IF NOT EXISTS idx_orgs_email
    ON public.orgs USING btree
    (email);

CREATE INDEX IF NOT EXISTS idx_orgs_phone
    ON public.orgs USING btree
    (phone);

--
-- Org Members
--

CREATE TABLE IF NOT EXISTS public.orgmembers
(
    id bigint NOT NULL,
    org_id bigint NOT NULL REFERENCES public.orgs ON DELETE RESTRICT,
    user_id bigint NOT NULL REFERENCES public.users ON DELETE RESTRICT,
    role character varying(64) COLLATE pg_catalog."default",
    source character varying(16) COLLATE pg_catalog."default",
    CONSTRAINT orgmembers_pkey PRIMARY KEY (id),
    CONSTRAINT membership UNIQUE (org_id, user_id)
)
WITH (
    OIDS = FALSE
);

ALTER TABLE public.orgmembers
    OWNER to postgres;

CREATE INDEX IF NOT EXISTS idx_orgmembers_team_id
    ON public.orgmembers USING btree
    (org_id ASC NULLS LAST);

CREATE INDEX IF NOT EXISTS idx_orgmembers_user_id
    ON public.orgmembers USING btree
    (user_id ASC NULLS LAST);

--
-- REPOS
--
CREATE TABLE IF NOT EXISTS public.repos
(
    id bigint NOT NULL,
    org_id bigint NOT NULL,
    extern_id bigint NOT NULL,
    provider character varying(16) COLLATE pg_catalog."default" NOT NULL,
    name character varying(64) COLLATE pg_catalog."default" NOT NULL,
    email character varying(160) COLLATE pg_catalog."default" NOT NULL,
    company character varying(64) COLLATE pg_catalog."default" NULL,
    avatar_url character varying(256) COLLATE pg_catalog."default" NULL,
    type character varying(16) COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    CONSTRAINT repos_pkey PRIMARY KEY (id),
    CONSTRAINT repos_provider_id UNIQUE (extern_id, provider),
    CONSTRAINT repos_org_name UNIQUE (org_id, name)
)
WITH (
    OIDS = FALSE
);

ALTER TABLE public.repos
    OWNER to postgres;

CREATE INDEX IF NOT EXISTS idx_repos_orgid
    ON public.repos USING btree
    (org_id)
    ;

CREATE INDEX IF NOT EXISTS idx_repos_provider
    ON public.repos USING btree
    (provider);

--
--
--
COMMIT;

