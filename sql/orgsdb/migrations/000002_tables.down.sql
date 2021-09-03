BEGIN;

DROP INDEX IF EXISTS idx_orgs_extern_id;
DROP INDEX IF EXISTS idx_orgs_reg_id;
DROP INDEX IF EXISTS idx_orgs_approver_email;
DROP INDEX IF EXISTS idx_orgs_expires_at;

--
--
--
COMMIT;
