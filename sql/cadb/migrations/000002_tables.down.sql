BEGIN;

DROP TABLE IF EXISTS public.issuers;
DROP INDEX IF EXISTS idx_issuers_label

DROP TABLE IF EXISTS public.cert_profiles;
DROP INDEX IF EXISTS idx_cert_profiles_label

--
--
--
COMMIT;
