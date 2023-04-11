BEGIN;

-- Remove SKID UNIQUE constraint
DROP INDEX IF EXISTS idx_certificates_skid;

CREATE INDEX IF NOT EXISTS idx_certificates_skid
    ON public.certificates USING btree
    (skid COLLATE pg_catalog."default");


-- Remove SKID UNIQUE constraint
ALTER TABLE public.revoked DROP CONSTRAINT unique_revoked_skid;

DROP INDEX IF EXISTS unique_revoked_skid;
DROP INDEX IF EXISTS idx_revoked_skid;

CREATE INDEX IF NOT EXISTS idx_revoked_skid
    ON public.revoked USING btree
    (skid COLLATE pg_catalog."default");

--
--
--
COMMIT;
