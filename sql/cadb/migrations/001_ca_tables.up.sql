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
    issuers_pem text COLLATE pg_catalog."default" NOT NULL,
    profile character varying(32) COLLATE pg_catalog."default" NOT NULL,
    locations text COLLATE pg_catalog."default" NOT NULL,
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
    issuers_pem text COLLATE pg_catalog."default" NOT NULL,
    profile character varying(32) COLLATE pg_catalog."default" NOT NULL,
    revoked_at timestamp with time zone,
    reason int NOT NULL,
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

CREATE TABLE IF NOT EXISTS public.crls
(
    id bigint NOT NULL,
    ikid character varying(64) COLLATE pg_catalog."default" NOT NULL,
    this_update timestamp with time zone,
    next_update timestamp with time zone,
    issuer character varying(260) COLLATE pg_catalog."default" NOT NULL,
    pem text COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT crls_pkey PRIMARY KEY (id),
    CONSTRAINT crls_ikid UNIQUE (ikid)
)
WITH (
    OIDS = FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_crls_ikid
    ON public.crls USING btree
    (ikid COLLATE pg_catalog."default");

CREATE INDEX IF NOT EXISTS idx_crls_next_update
    ON public.crls USING btree
    (next_update);

SELECT create_constraint_if_not_exists(
    'public',
    'crls',
    'idx_crls_ikid',
    'ALTER TABLE public.crls ADD CONSTRAINT unique_crls_ikid UNIQUE USING INDEX idx_crls_ikid;');

--
--
--
COMMIT;
