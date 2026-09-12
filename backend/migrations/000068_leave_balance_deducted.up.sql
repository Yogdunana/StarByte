-- ============================================================
-- 000068_leave_balance_deducted.up.sql
-- 记录提交时是否预扣余额，驳回按快照返还，避免中途改 Deductible 错账
-- ============================================================

ALTER TABLE leave_applications
    ADD COLUMN IF NOT EXISTS balance_deducted BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN leave_applications.balance_deducted IS '提交时是否已预扣年假/调休；驳回按此快照返还，不读当前类型 deductible';

UPDATE leave_applications a
SET balance_deducted = TRUE
FROM leave_types t
WHERE a.leave_type_id = t.id
  AND t.deductible = TRUE
  AND a.status = 'pending'
  AND a.deleted_at IS NULL
  AND a.balance_deducted = FALSE;
