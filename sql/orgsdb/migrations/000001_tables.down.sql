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

DROP TABLE IF EXISTS public.fcc_frn;
DROP INDEX IF EXISTS idx_fcc_frn_updated_at;

DROP TABLE IF EXISTS public.fcc_contact;
DROP INDEX IF EXISTS idx_fcc_contact_frn;
DROP INDEX IF EXISTS idx_fcc_contact_updated_at;

DROP TABLE IF EXISTS public.orgtokens;
DROP INDEX IF EXISTS idx_orgtokens_token_code;
DROP INDEX IF EXISTS idx_orgtokens_org_id;

DROP TABLE IF EXISTS apikeys;
DROP TABLE IF EXISTS idx_apikeys_key

DROP TABLE IF EXISTS subscriptions;
DROP INDEX IF EXISTS idx_subscriptions_id_user_id;
DROP INDEX IF EXISTS idx_subscriptions_external_id;

COMMIT;
