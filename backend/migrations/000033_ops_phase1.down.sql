-- ============================================================
-- 000033_ops_phase1.down.sql
-- ============================================================

DROP INDEX IF EXISTS idx_contracts_expired_at;
DROP INDEX IF EXISTS idx_discipline_records_level;

ALTER TABLE contracts
    DROP COLUMN IF EXISTS start_at,
    DROP COLUMN IF EXISTS amount,
    DROP COLUMN IF EXISTS party_name,
    DROP COLUMN IF EXISTS contract_type;

ALTER TABLE discipline_records
    DROP COLUMN IF EXISTS revoked_at,
    DROP COLUMN IF EXISTS revoked_by,
    DROP COLUMN IF EXISTS revoke_reason,
    DROP COLUMN IF EXISTS approved_at,
    DROP COLUMN IF EXISTS approved_by,
    DROP COLUMN IF EXISTS flow_instance_id;
