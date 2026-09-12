-- ============================================================
-- 000055_smtp_settings.down.sql
-- ============================================================

DELETE FROM configs WHERE config_key = 'smtp_settings';
