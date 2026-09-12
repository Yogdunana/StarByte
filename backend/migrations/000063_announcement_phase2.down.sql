-- ============================================================
-- 000063_announcement_phase2.down.sql
-- ============================================================

DELETE FROM scheduler_tasks WHERE code = 'announcement_auto_unpublish';

UPDATE notification_templates
SET channels = '["in_app","websocket"]'
WHERE code = 'announcement_published' AND deleted_at IS NULL;

ALTER TABLE announcement_reads DROP COLUMN IF EXISTS duration_seconds;

DROP INDEX IF EXISTS idx_announcements_sort;
DROP INDEX IF EXISTS idx_announcements_expires;

ALTER TABLE announcements DROP CONSTRAINT IF EXISTS chk_announcements_audience_type;

ALTER TABLE announcements
    DROP COLUMN IF EXISTS attachments,
    DROP COLUMN IF EXISTS audience_ids,
    DROP COLUMN IF EXISTS audience_type,
    DROP COLUMN IF EXISTS sort_order,
    DROP COLUMN IF EXISTS expires_at;
