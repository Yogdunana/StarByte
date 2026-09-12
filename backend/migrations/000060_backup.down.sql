DELETE FROM scheduler_tasks
WHERE code IN ('backup_scheduled_full', 'backup_retention_cleanup')
  AND status <> 2;

DELETE FROM notification_templates WHERE code = 'backup_failed';

DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE code IN ('backup:read', 'backup:create', 'backup:delete', 'backup:restore', 'backup:manage')
);

DELETE FROM permissions
WHERE code IN ('backup:read', 'backup:create', 'backup:delete', 'backup:restore', 'backup:manage');

DROP TABLE IF EXISTS backup_records;
DROP TABLE IF EXISTS backup_policies;
