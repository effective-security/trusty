BEGIN;

DROP TABLE IF EXISTS public.fcc_frn;
DROP INDEX IF EXISTS idx_fcc_frn_updated_at;

DROP TABLE IF EXISTS public.fcc_contact;
DROP INDEX IF EXISTS idx_fcc_contact_frn;
DROP INDEX IF EXISTS idx_fcc_contact_updated_at;

COMMIT;
