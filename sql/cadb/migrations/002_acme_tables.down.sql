BEGIN;

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

--
--
--
COMMIT;
