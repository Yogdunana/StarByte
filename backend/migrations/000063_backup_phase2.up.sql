-- ============================================================
-- 000063_backup_phase2.up.sql
-- Issue #88 phase-2：AES 标记、失败告警补邮件渠道
-- ============================================================

ALTER TABLE backup_records
    ADD COLUMN IF NOT EXISTS encrypted BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE notification_templates
SET channels = '["in_app","websocket","email"]'
WHERE code = 'backup_failed';
