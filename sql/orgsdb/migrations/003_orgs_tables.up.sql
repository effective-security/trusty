BEGIN;

ALTER TABLE public.orgs
ADD COLUMN IF NOT EXISTS reg_id character varying(32) COLLATE pg_catalog."default" NOT NULL;

--
--
--
COMMIT;
