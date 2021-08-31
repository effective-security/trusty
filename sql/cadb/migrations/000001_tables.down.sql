BEGIN;

DROP TABLE IF EXISTS public.certificates;
DROP INDEX IF EXISTS unique_certificates_skid;
DROP INDEX IF EXISTS unique_certificates_sha256;
DROP INDEX IF EXISTS idx_certificates_org;
DROP INDEX IF EXISTS idx_certificates_skid;
DROP INDEX IF EXISTS idx_certificates_ikid;
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

DROP TABLE IF EXISTS public.registrations;
--DROP INDEX IF EXISTS unique_registrations_external_id;
--DROP INDEX IF EXISTS registrations_external_id;
DROP INDEX IF EXISTS unique_registrations_key_id;
DROP INDEX IF EXISTS registrations_key_id;
DROP INDEX IF EXISTS idx_registrations_created_at;

DROP TABLE IF EXISTS public.orders;
DROP INDEX IF EXISTS idx_orders_reg_id;
DROP INDEX IF EXISTS idx_orders_names_hash;

DROP TABLE IF EXISTS public.acmecerts;
DROP INDEX IF EXISTS idx_acmecerts_reg_id;
DROP INDEX IF EXISTS idx_acmecerts_order_id;

DROP TABLE IF EXISTS public.authorizations;
DROP INDEX IF EXISTS idx_authorizations_reg_id;

DROP TABLE IF EXISTS nonces;
DROP INDEX IF EXISTS idx_nonces_nonce;
DROP INDEX IF EXISTS idx_nonces_expires_at;

--
--
--
COMMIT;
