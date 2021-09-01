BEGIN;

ALTER TABLE public.acmecerts
ADD COLUMN IF NOT EXISTS locations text COLLATE pg_catalog."default" NULL;

--
--
--
COMMIT;
