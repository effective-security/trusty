BEGIN;

DROP TABLE IF EXISTS public.users;
DROP INDEX IF EXISTS unique_users_email;
DROP INDEX IF EXISTS unique_users_login;

DROP TABLE IF EXISTS public.orgs;
DROP INDEX IF EXISTS unique_orgs_name;
DROP INDEX IF EXISTS unique_orgs_login;
DROP INDEX IF EXISTS idx_orgs_provider;

DROP TABLE IF EXISTS public.repos;
DROP INDEX IF EXISTS idx_repos_orgid;
DROP INDEX IF EXISTS idx_repos_provider;

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

COMMIT;
