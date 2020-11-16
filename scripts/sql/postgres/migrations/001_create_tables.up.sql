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

CREATE TABLE IF NOT EXISTS public.users
(
    id bigint NOT NULL,
    extern_id bigint NULL,
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
    CONSTRAINT users_provider_id UNIQUE (extern_id, provider)
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

CREATE TABLE IF NOT EXISTS public.certificates
(
    id bigint NOT NULL,
    owner_id bigint NOT NULL,
    skid character varying(64) COLLATE pg_catalog."default" NOT NULL,
    ikid character varying(64) COLLATE pg_catalog."default" NOT NULL,
    sn character varying(32) COLLATE pg_catalog."default" NOT NULL,
    notbefore timestamp with time zone,
    notafter timestamp with time zone,
    subject character varying(260) COLLATE pg_catalog."default" NOT NULL,
    pem text COLLATE pg_catalog."default" NOT NULL,
    profile character varying(32) COLLATE pg_catalog."default" NULL,
    role character varying(32) COLLATE pg_catalog."default" NULL,
    host character varying(160) COLLATE pg_catalog."default" NULL,
    CONSTRAINT certificates_pkey PRIMARY KEY (id),
    CONSTRAINT certificates_issuer_sn UNIQUE (ikid, sn)
)
WITH (
    OIDS = FALSE
);

CREATE INDEX IF NOT EXISTS idx_certificates_owner
    ON public.certificates USING btree
    (owner_id);

CREATE INDEX IF NOT EXISTS idx_certificates_skid
    ON public.certificates USING btree
    (skid COLLATE pg_catalog."default");

CREATE INDEX IF NOT EXISTS idx_certificates_ikid
    ON public.certificates USING btree
    (ikid COLLATE pg_catalog."default");

CREATE INDEX IF NOT EXISTS idx_certificates_notafter
    ON public.certificates USING btree
    (notafter);

CREATE TABLE IF NOT EXISTS public.revoked
(
    id bigint NOT NULL,
    owner_id bigint NOT NULL,
    skid character varying(64) COLLATE pg_catalog."default" NOT NULL,
    ikid character varying(64) COLLATE pg_catalog."default" NOT NULL,
    sn character varying(32) COLLATE pg_catalog."default" NOT NULL,
    notbefore timestamp with time zone,
    notafter timestamp with time zone,
    subject character varying(260) COLLATE pg_catalog."default" NOT NULL,
    pem text COLLATE pg_catalog."default" NOT NULL,
    profile character varying(32) COLLATE pg_catalog."default" NULL,
    role character varying(32) COLLATE pg_catalog."default" NULL,
    host character varying(160) COLLATE pg_catalog."default" NULL,
    revoked_at timestamp with time zone,
    reason character varying(254) COLLATE pg_catalog."default" NULL,
    requestor character varying(160) COLLATE pg_catalog."default" NULL,
    CONSTRAINT revoked_pkey PRIMARY KEY (id),
    CONSTRAINT revoked_issuer_sn UNIQUE (ikid, sn)
)
WITH (
    OIDS = FALSE
);

CREATE INDEX IF NOT EXISTS idx_revoked_owner
    ON public.revoked USING btree
    (owner_id);

CREATE INDEX IF NOT EXISTS idx_revoked_skid
    ON public.revoked USING btree
    (skid COLLATE pg_catalog."default");

CREATE INDEX IF NOT EXISTS idx_revoked_ikid
    ON public.revoked USING btree
    (ikid COLLATE pg_catalog."default");

CREATE INDEX IF NOT EXISTS idx_revoked_notafter
    ON public.revoked USING btree
    (notafter);

COMMIT;

