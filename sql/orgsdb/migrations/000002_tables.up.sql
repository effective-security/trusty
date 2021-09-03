BEGIN;

CREATE INDEX IF NOT EXISTS idx_orgs_extern_id
    ON public.orgs USING btree
    (extern_id);

CREATE INDEX IF NOT EXISTS idx_orgs_reg_id
    ON public.orgs USING btree
    (reg_id);

CREATE INDEX IF NOT EXISTS idx_orgs_approver_email
    ON public.orgs USING btree
    (approver_email);

CREATE INDEX IF NOT EXISTS idx_orgs_expires_at
    ON public.orgs USING btree
    (expires_at);


--
--
--
COMMIT;
