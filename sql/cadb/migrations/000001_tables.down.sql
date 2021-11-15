BEGIN;

DROP TABLE IF EXISTS public.certificates;
DROP INDEX IF EXISTS idx_certificates_org;
DROP INDEX IF EXISTS idx_certificates_skid;
DROP INDEX IF EXISTS idx_certificates_ikid;
DROP INDEX IF EXISTS idx_certificates_ikid_serial;
DROP INDEX IF EXISTS idx_certificates_notafter;
DROP INDEX IF EXISTS idx_certificates_sha256;

DROP TABLE IF EXISTS public.revoked;
DROP INDEX IF EXISTS unique_revoked_skid;
DROP INDEX IF EXISTS unique_revoked_sha256;
DROP INDEX IF EXISTS idx_revoked_org;
DROP INDEX IF EXISTS idx_revoked_skid;
DROP INDEX IF EXISTS idx_revoked_ikid;
DROP INDEX IF EXISTS idx_revoked_notafter;
DROP INDEX IF EXISTS idx_revoked_sha256;
DROP INDEX IF EXISTS idx_revoked_sn;

DROP TABLE IF EXISTS public.roots;
DROP INDEX IF EXISTS unique_roots_skid;
DROP INDEX IF EXISTS unique_roots_sha256;
DROP INDEX IF EXISTS idx_roots_skid;
DROP INDEX IF EXISTS idx_roots_notafter;
DROP INDEX IF EXISTS idx_roots_sha256;

DROP TABLE IF EXISTS public.crls;
DROP INDEX IF EXISTS unique_crls_ikid;
DROP INDEX IF EXISTS idx_crls_ikid;
DROP INDEX IF EXISTS idx_crls_next_update;

DROP TABLE IF EXISTS nonces;
DROP INDEX IF EXISTS idx_nonces_nonce;
DROP INDEX IF EXISTS idx_nonces_expires_at;

DROP TABLE IF EXISTS public.issuers;
DROP INDEX IF EXISTS idx_issuers_label

DROP TABLE IF EXISTS public.cert_profiles;
DROP INDEX IF EXISTS idx_cert_profiles_label

--
--
--
COMMIT;
