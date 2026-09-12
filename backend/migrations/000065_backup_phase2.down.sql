-- ============================================================
-- 000065_backup_phase2.down.sql
-- ============================================================

UPDATE notification_templates
SET channels = '["in_app","websocket"]'
WHERE code = 'backup_failed';

ALTER TABLE backup_records
    DROP COLUMN IF EXISTS encrypted;
