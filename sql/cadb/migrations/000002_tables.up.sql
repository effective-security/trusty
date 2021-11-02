BEGIN;

ALTER TABLE public.acmecerts
ADD COLUMN IF NOT EXISTS locations text COLLATE pg_catalog."default" NULL;

CREATE INDEX IF NOT EXISTS idx_revoked_sn
    ON public.revoked USING btree
    (serial_number COLLATE pg_catalog."default");

--
--
--
COMMIT;
