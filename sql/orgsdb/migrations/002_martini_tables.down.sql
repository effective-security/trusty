BEGIN;

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
