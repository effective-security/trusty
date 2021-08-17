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
-- FRN
--
CREATE TABLE IF NOT EXISTS public.fcc_frn
(
    filer_id bigint NOT NULL,
    json text COLLATE pg_catalog."default" NULL,
    updated_at timestamp with time zone,
    CONSTRAINT fcc_frn_pkey PRIMARY KEY (filer_id)
)
WITH (
    OIDS = FALSE
);

CREATE INDEX IF NOT EXISTS idx_fcc_frn_updated_at
    ON public.fcc_frn USING btree
    (updated_at);

--
-- FRN Contact
--
CREATE TABLE IF NOT EXISTS public.fcc_contact
(
    id bigint NOT NULL,
    frn character varying(16) COLLATE pg_catalog."default" NOT NULL,
    json text COLLATE pg_catalog."default" NOT NULL,
    updated_at timestamp with time zone,
    CONSTRAINT fcc_contact_pkey PRIMARY KEY (id),
    CONSTRAINT fcc_contact_frn UNIQUE (frn)
)
WITH (
    OIDS = FALSE
);

CREATE INDEX IF NOT EXISTS idx_fcc_contact_updated_at
    ON public.fcc_contact USING btree
    (updated_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fcc_contact_frn
    ON public.fcc_contact USING btree
    (frn);

--
-- Org Tokens
--
CREATE TABLE IF NOT EXISTS public.orgtokens
(
    id bigint NOT NULL,
    org_id bigint NOT NULL,
    requestor_id bigint NOT NULL,
    approver_email character varying(160) COLLATE pg_catalog."default" NOT NULL,
    token character varying(16) COLLATE pg_catalog."default" NOT NULL,
    code character varying(6) COLLATE pg_catalog."default" NOT NULL,
    used boolean NOT NULL,
    created_at timestamp with time zone,
    expires_at timestamp with time zone,
    used_at timestamp with time zone,
    CONSTRAINT orgtokens_pkey PRIMARY KEY (id),
    CONSTRAINT orgtokens_token_code UNIQUE (token, code)
)
WITH (
    OIDS = FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_orgtokens_token_code
    ON public.orgtokens USING btree
    (token,code);

CREATE INDEX IF NOT EXISTS idx_orgtokens_org_id
    ON public.orgtokens USING btree
    (org_id);

--
-- API Keys
--
CREATE TABLE IF NOT EXISTS public.apikeys
(
    id bigint NOT NULL,
    org_id bigint NOT NULL,
    key character varying(64) COLLATE pg_catalog."default" NOT NULL,
    enrollment boolean NOT NULL,
    management boolean NOT NULL,
    billing boolean NOT NULL,
    created_at timestamp with time zone,
    expires_at timestamp with time zone,
    used_at timestamp with time zone,
    CONSTRAINT apikeys_pkey PRIMARY KEY (id),
    CONSTRAINT apikeys_key UNIQUE (key)
)
WITH (
    OIDS = FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_apikeys_key
    ON public.apikeys USING btree
    (key);
--
--  Subscriptions
--
CREATE TABLE IF NOT EXISTS public.subscriptions
(
    id bigint NOT NULL,
    external_id character varying(32) COLLATE pg_catalog."default" NOT NULL,
    user_id bigint NOT NULL,
    customer_id character varying(32) COLLATE pg_catalog."default" NOT NULL,
    price_id character varying(32) COLLATE pg_catalog."default" NOT NULL,
    price_amount bigint NOT NULL,
    price_currency character varying(32) COLLATE pg_catalog."default" NOT NULL,
    payment_method_id character varying(32) COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone,
    expires_at timestamp with time zone,
    last_paid_at timestamp with time zone,
    status character varying(32) COLLATE pg_catalog."default" NOT NULL
)
WITH (
    OIDS = FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_id_user_id
    ON public.subscriptions USING btree
    (id,user_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_external_id
    ON public.subscriptions USING btree
    (external_id);
--
--
--
COMMIT;
