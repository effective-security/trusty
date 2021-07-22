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
--
--
COMMIT;
