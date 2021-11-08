BEGIN;

--
-- Issuers
--
CREATE TABLE IF NOT EXISTS public.issuers
(
    id bigint NOT NULL,
    label character varying(32) COLLATE pg_catalog."default" NOT NULL,
    config text COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone DEFAULT Now(),
    updated_at timestamp with time zone DEFAULT Now(),
    CONSTRAINT issuers_pkey PRIMARY KEY (id),
    CONSTRAINT issuers_label UNIQUE (label)
)
WITH (
    OIDS = FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_issuers_label
    ON public.issuers USING btree
    (label COLLATE pg_catalog."default");

SELECT create_constraint_if_not_exists(
    'public',
    'issuers',
    'unique_issuers_label',
    'ALTER TABLE public.issuers ADD CONSTRAINT unique_issuers_label UNIQUE USING INDEX idx_issuers_label;');


--
-- Certificate profiles
--
CREATE TABLE IF NOT EXISTS public.cert_profiles
(
    id bigint NOT NULL,
    label character varying(32) COLLATE pg_catalog."default" NOT NULL,
    issuer_label character varying(32) COLLATE pg_catalog."default" NOT NULL,
    config text COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone DEFAULT Now(),
    updated_at timestamp with time zone DEFAULT Now(),
    CONSTRAINT cert_profiles_pkey PRIMARY KEY (id),
    CONSTRAINT cert_profiles_label UNIQUE (label)
)
WITH (
    OIDS = FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_cert_profiles_label
    ON public.cert_profiles USING btree
    (label COLLATE pg_catalog."default");

SELECT create_constraint_if_not_exists(
    'public',
    'cert_profiles',
    'unique_cert_profiles_label',
    'ALTER TABLE public.cert_profiles ADD CONSTRAINT unique_cert_profiles_label UNIQUE USING INDEX idx_cert_profiles_label;');


--
--
--
COMMIT;
