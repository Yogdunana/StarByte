-- ============================================================
-- 000056_announcement.down.sql
-- ============================================================

DELETE FROM scheduler_tasks WHERE code = 'announcement_scheduled_publish';

DELETE FROM notification_templates WHERE code = 'announcement_published' AND deleted_at IS NULL;

DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE code IN (
        'announcement:read', 'announcement:create', 'announcement:update',
        'announcement:delete', 'announcement:publish', 'announcement:manage'
    )
);

DELETE FROM permissions
WHERE code IN (
    'announcement:read', 'announcement:create', 'announcement:update',
    'announcement:delete', 'announcement:publish', 'announcement:manage'
);

DROP TABLE IF EXISTS announcement_reads;
DROP TABLE IF EXISTS announcements;
