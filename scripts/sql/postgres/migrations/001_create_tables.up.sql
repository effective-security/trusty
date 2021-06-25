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
    extern_id bigint NOT NULL,
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
    CONSTRAINT users_pkey PRIMARY KEY (id)
)
WITH (
    OIDS = FALSE
);

ALTER TABLE public.users
    OWNER to postgres;

CREATE UNIQUE INDEX IF NOT EXISTS unique_users_email
    ON public.users USING btree
    (email COLLATE pg_catalog."default")
    ;

CREATE UNIQUE INDEX IF NOT EXISTS unique_users_login
    ON public.users USING btree
    (login COLLATE pg_catalog."default")
    ;

SELECT create_constraint_if_not_exists(
    'public',
    'users',
    'unique_users_email',
    'ALTER TABLE public.users ADD CONSTRAINT unique_users_email UNIQUE USING INDEX unique_users_email;');

SELECT create_constraint_if_not_exists(
    'public',
    'users',
    'unique_users_login',
    'ALTER TABLE public.users ADD CONSTRAINT unique_users_login UNIQUE USING INDEX unique_users_login;');

--
-- ORGANIZATIONS
--
CREATE TABLE IF NOT EXISTS public.orgs
(
    id bigint NOT NULL,
    extern_id bigint NOT NULL,
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
-- CERTIFICATES
--
CREATE TABLE IF NOT EXISTS public.certificates
(
    id bigint NOT NULL,
    org_id bigint NOT NULL,
    skid character varying(64) COLLATE pg_catalog."default" NOT NULL,
    ikid character varying(64) COLLATE pg_catalog."default" NOT NULL,
    serial_number character varying(64) COLLATE pg_catalog."default" NOT NULL,
    not_before timestamp with time zone,
    no_tafter timestamp with time zone,
    subject character varying(260) COLLATE pg_catalog."default" NOT NULL,
    issuer character varying(260) COLLATE pg_catalog."default" NOT NULL,
    sha256 character varying(64) COLLATE pg_catalog."default" NOT NULL,
    pem text COLLATE pg_catalog."default" NOT NULL,
    issuers_pem text COLLATE pg_catalog."default" NULL,
    profile character varying(32) COLLATE pg_catalog."default" NULL,
    CONSTRAINT certificates_pkey PRIMARY KEY (id),
    CONSTRAINT certificates_skid UNIQUE (skid),
    CONSTRAINT certificates_sha256 UNIQUE (sha256),
    CONSTRAINT certificates_issuer_sn UNIQUE (ikid, serial_number)
)
WITH (
    OIDS = FALSE
);

CREATE INDEX IF NOT EXISTS idx_certificates_org
    ON public.certificates USING btree
    (org_id);

CREATE INDEX IF NOT EXISTS idx_certificates_ikid
    ON public.certificates USING btree
    (ikid COLLATE pg_catalog."default");

CREATE INDEX IF NOT EXISTS idx_certificates_notafter
    ON public.certificates USING btree
    (no_tafter);

CREATE UNIQUE INDEX IF NOT EXISTS idx_certificates_sha256
    ON public.certificates USING btree
    (sha256 COLLATE pg_catalog."default");

CREATE UNIQUE INDEX IF NOT EXISTS idx_certificates_skid
    ON public.certificates USING btree
    (skid COLLATE pg_catalog."default");

SELECT create_constraint_if_not_exists(
    'public',
    'certificates',
    'unique_certificates_skid',
    'ALTER TABLE public.certificates ADD CONSTRAINT unique_certificates_skid UNIQUE USING INDEX idx_certificates_skid;');

SELECT create_constraint_if_not_exists(
    'public',
    'certificates',
    'unique_certificates_sha256',
    'ALTER TABLE public.certificates ADD CONSTRAINT unique_certificates_sha256 UNIQUE USING INDEX idx_certificates_sha256;');

--
-- REVOKED CERTIFICATES
--
CREATE TABLE IF NOT EXISTS public.revoked
(
    id bigint NOT NULL,
    org_id bigint NOT NULL,
    skid character varying(64) COLLATE pg_catalog."default" NOT NULL,
    ikid character varying(64) COLLATE pg_catalog."default" NOT NULL,
    serial_number character varying(64) COLLATE pg_catalog."default" NOT NULL,
    not_before timestamp with time zone,
    no_tafter timestamp with time zone,
    subject character varying(260) COLLATE pg_catalog."default" NOT NULL,
    issuer character varying(260) COLLATE pg_catalog."default" NOT NULL,
    sha256 character varying(64) COLLATE pg_catalog."default" NOT NULL,
    pem text COLLATE pg_catalog."default" NOT NULL,
    issuers_pem text COLLATE pg_catalog."default" NULL,
    profile character varying(32) COLLATE pg_catalog."default" NULL,
    revoked_at timestamp with time zone,
    reason int NULL,
    CONSTRAINT revoked_pkey PRIMARY KEY (id),
    CONSTRAINT revoked_skid UNIQUE (skid),
    CONSTRAINT revoked_sha256 UNIQUE (sha256),
    CONSTRAINT revoked_issuer_sn UNIQUE (ikid, serial_number)
)
WITH (
    OIDS = FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_revoked_skid
    ON public.revoked USING btree
    (skid COLLATE pg_catalog."default");

CREATE UNIQUE INDEX IF NOT EXISTS idx_revoked_sha256
    ON public.revoked USING btree
    (sha256 COLLATE pg_catalog."default");

CREATE INDEX IF NOT EXISTS idx_revoked_org
    ON public.revoked USING btree
    (org_id);

CREATE INDEX IF NOT EXISTS idx_revoked_ikid
    ON public.revoked USING btree
    (ikid COLLATE pg_catalog."default");

CREATE INDEX IF NOT EXISTS idx_revoked_notafter
    ON public.revoked USING btree
    (no_tafter);

SELECT create_constraint_if_not_exists(
    'public',
    'revoked',
    'unique_revoked_skid',
    'ALTER TABLE public.revoked ADD CONSTRAINT unique_revoked_skid UNIQUE USING INDEX idx_revoked_skid;');

SELECT create_constraint_if_not_exists(
    'public',
    'revoked',
    'unique_revoked_sha256',
    'ALTER TABLE public.revoked ADD CONSTRAINT unique_revoked_sha256 UNIQUE USING INDEX idx_revoked_sha256;');

--
-- Authorities
--
CREATE TABLE IF NOT EXISTS public.roots
(
    id bigint NOT NULL,
    skid character varying(64) COLLATE pg_catalog."default" NOT NULL,
    not_before timestamp with time zone,
    no_tafter timestamp with time zone,
    subject character varying(260) COLLATE pg_catalog."default" NOT NULL,
    sha256 character varying(64) COLLATE pg_catalog."default" NOT NULL,
    trust int NOT NULL,
    pem text COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT roots_pkey PRIMARY KEY (id),
    CONSTRAINT roots_skid UNIQUE (skid),
    CONSTRAINT roots_sha256 UNIQUE (sha256)
)
WITH (
    OIDS = FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_roots_skid
    ON public.roots USING btree
    (skid COLLATE pg_catalog."default");

CREATE UNIQUE INDEX IF NOT EXISTS idx_roots_sha256
    ON public.roots USING btree
    (sha256 COLLATE pg_catalog."default");

CREATE INDEX IF NOT EXISTS idx_roots_notafter
    ON public.roots USING btree
    (no_tafter);

SELECT create_constraint_if_not_exists(
    'public',
    'roots',
    'unique_roots_skid',
    'ALTER TABLE public.roots ADD CONSTRAINT unique_roots_skid UNIQUE USING INDEX idx_roots_skid;');

SELECT create_constraint_if_not_exists(
    'public',
    'roots',
    'unique_roots_sha256',
    'ALTER TABLE public.roots ADD CONSTRAINT unique_roots_sha256 UNIQUE USING INDEX idx_roots_sha256;');

--
--
--
COMMIT;

