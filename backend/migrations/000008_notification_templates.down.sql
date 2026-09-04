-- ============================================================
-- 000008_notification_templates.down.sql
-- 回滚通知模板表
-- ============================================================

DROP INDEX IF EXISTS idx_notification_templates_deleted_at;
DROP INDEX IF EXISTS idx_notification_templates_status;
DROP INDEX IF EXISTS idx_notification_templates_code;
DROP TABLE IF EXISTS notification_templates;
