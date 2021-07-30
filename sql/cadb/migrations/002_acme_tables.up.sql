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
-- registrations
-- 
CREATE TABLE IF NOT EXISTS public.registrations
(
    id bigint NOT NULL,
    external_id character varying(64) COLLATE pg_catalog."default" NOT NULL,
    key_id character varying(64) COLLATE pg_catalog."default" NOT NULL,
    key text COLLATE pg_catalog."default" NOT NULL,
    contact text COLLATE pg_catalog."default" NULL,
    agreement text COLLATE pg_catalog."default" NULL,
    initial_ip character varying(64) COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone,
    status character varying(16) COLLATE pg_catalog."default" NOT NULL,

    CONSTRAINT registrations_pkey PRIMARY KEY (id),
    CONSTRAINT registrations_key_id UNIQUE (key_id)
    -- CONSTRAINT registrations_external_id UNIQUE (external_id)
)
WITH (
    OIDS = FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS registrations_key_id
    ON public.registrations USING btree
    (external_id COLLATE pg_catalog."default");

CREATE INDEX IF NOT EXISTS idx_registrations_created_at
    ON public.registrations USING btree
    (created_at);

SELECT create_constraint_if_not_exists(
    'public',
    'registrations',
    'registrations_key_id',
    'ALTER TABLE public.registrations ADD CONSTRAINT unique_registrations_key_id UNIQUE USING INDEX registrations_key_id;');

--CREATE UNIQUE INDEX IF NOT EXISTS registrations_external_id
--    ON public.registrations USING btree
--    (external_id COLLATE pg_catalog."default");

--SELECT create_constraint_if_not_exists(
--    'public',
--    'registrations',
--    'registrations_external_id',
--    'ALTER TABLE public.registrations ADD CONSTRAINT unique_registrations_external_id UNIQUE USING INDEX registrations_external_id;');

--
-- orders
-- 
CREATE TABLE IF NOT EXISTS public.orders
(
    id bigint NOT NULL,
    reg_id bigint NOT NULL,
    names_hash character varying(64) COLLATE pg_catalog."default" NOT NULL,
    created_at timestamp with time zone,
    status character varying(16) COLLATE pg_catalog."default" NOT NULL,
    expires_at timestamp with time zone,
    cert_id character varying(64) COLLATE pg_catalog."default" NOT NULL,
    binding_id character varying(64) COLLATE pg_catalog."default" NOT NULL,
    external_order_id bigint NOT NULL,
    json text COLLATE pg_catalog."default" NULL,

    CONSTRAINT orders_pkey PRIMARY KEY (id),
    CONSTRAINT orders_names_hash UNIQUE (reg_id,names_hash)
)
WITH (
    OIDS = FALSE
);

CREATE INDEX IF NOT EXISTS idx_orders_reg_id
    ON public.orders USING btree
    (reg_id);

CREATE INDEX IF NOT EXISTS idx_orders_names_hash
    ON public.orders USING btree
    (names_hash);

--
-- acmecerts
-- 
CREATE TABLE IF NOT EXISTS public.acmecerts
(
    id bigint NOT NULL,
    reg_id bigint NOT NULL,
    order_id bigint NOT NULL,
    binding_id character varying(64) COLLATE pg_catalog."default" NOT NULL,
    external_id bigint NOT NULL,
    pem text COLLATE pg_catalog."default" NOT NULL,

    CONSTRAINT acmecerts_pkey PRIMARY KEY (id),
    CONSTRAINT acmecerts_names_hash UNIQUE (reg_id,order_id)
)
WITH (
    OIDS = FALSE
);

CREATE INDEX IF NOT EXISTS idx_acmecerts_reg_id
    ON public.acmecerts USING btree
    (reg_id);

CREATE INDEX IF NOT EXISTS idx_acmecerts_order_id
    ON public.acmecerts USING btree
    (order_id);

--
-- authorizations
-- 
CREATE TABLE IF NOT EXISTS public.authorizations
(
    id bigint NOT NULL,
    reg_id bigint NOT NULL,
    type character varying(64) COLLATE pg_catalog."default" NOT NULL,
    value character varying(64) COLLATE pg_catalog."default" NOT NULL,
    status character varying(16) COLLATE pg_catalog."default" NOT NULL,
    expires_at timestamp with time zone,
    challenges text COLLATE pg_catalog."default" NOT NULL,

    CONSTRAINT authorizations_pkey PRIMARY KEY (id)
)
WITH (
    OIDS = FALSE
);

CREATE INDEX IF NOT EXISTS idx_authorizations_reg_id
    ON public.authorizations USING btree
    (reg_id);


--
-- Nonces
--
CREATE TABLE IF NOT EXISTS public.nonces
(
    id bigint NOT NULL,
    nonce character varying(16) COLLATE pg_catalog."default" NOT NULL,
    used boolean NOT NULL,
    created_at timestamp with time zone,
    expires_at timestamp with time zone,
    used_at timestamp with time zone,
    CONSTRAINT nonces_pkey PRIMARY KEY (id),
    CONSTRAINT nonces_nonce UNIQUE (nonce)
)
WITH (
    OIDS = FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_nonces_nonce
    ON public.nonces USING btree
    (nonce);

CREATE UNIQUE INDEX IF NOT EXISTS idx_nonces_expires_at
    ON public.nonces USING btree
    (expires_at);

--
--
--
COMMIT;
