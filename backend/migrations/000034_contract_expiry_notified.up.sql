-- ============================================================
-- 000034_contract_expiry_notified.up.sql
-- 合同到期提醒只发一次，避免每日任务重复通知
-- ============================================================

ALTER TABLE contracts
    ADD COLUMN IF NOT EXISTS expiry_notified_at TIMESTAMP;
