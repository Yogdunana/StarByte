DROP INDEX IF EXISTS idx_task_logs_created_at;
DROP INDEX IF EXISTS idx_task_logs_task_id;
DROP TABLE IF EXISTS task_logs;

ALTER TABLE task_attachments
    DROP COLUMN IF EXISTS uploaded_by;

ALTER TABLE task_comments
    DROP COLUMN IF EXISTS mentions,
    DROP COLUMN IF EXISTS updated_at;

DROP INDEX IF EXISTS idx_tasks_due_date;
DROP INDEX IF EXISTS idx_tasks_parent_id;

ALTER TABLE tasks
    DROP COLUMN IF EXISTS parent_id,
    DROP COLUMN IF EXISTS sort_order,
    DROP COLUMN IF EXISTS due_reminded_at,
    DROP COLUMN IF EXISTS overdue_reminded_at;

DELETE FROM permissions WHERE code IN ('task:assign', 'task:transfer', 'task:comment');
