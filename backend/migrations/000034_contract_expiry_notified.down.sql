-- ============================================================
-- 000034_contract_expiry_notified.down.sql
-- ============================================================

ALTER TABLE contracts
    DROP COLUMN IF EXISTS expiry_notified_at;
