BEGIN;

DROP TABLE IF EXISTS public.users;
DROP INDEX IF EXISTS idx_users_provider_email;
DROP INDEX IF EXISTS idx_users_provider_login;
DROP INDEX IF EXISTS idx_users_extern_id;
DROP INDEX IF EXISTS unique_users_provider_login;
DROP INDEX IF EXISTS unique_users_provider_email;

DROP TABLE IF EXISTS public.orgs;
DROP INDEX IF EXISTS unique_orgs_name;
DROP INDEX IF EXISTS unique_orgs_login;
DROP INDEX IF EXISTS idx_orgs_provider;
DROP INDEX IF EXISTS idx_orgs_email;
DROP INDEX IF EXISTS idx_orgs_phone;

DROP TABLE IF EXISTS public.repos;
DROP INDEX IF EXISTS idx_repos_orgid;
DROP INDEX IF EXISTS idx_repos_provider;

COMMIT;
