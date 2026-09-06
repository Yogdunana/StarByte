-- Soft-deleted scheduler task codes may be reused.
DROP INDEX IF EXISTS idx_scheduler_tasks_code;
CREATE UNIQUE INDEX IF NOT EXISTS idx_scheduler_tasks_code_active
    ON scheduler_tasks (code) WHERE status <> 2;
