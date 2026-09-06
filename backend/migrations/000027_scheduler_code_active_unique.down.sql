DROP INDEX IF EXISTS idx_scheduler_tasks_code_active;
CREATE UNIQUE INDEX IF NOT EXISTS idx_scheduler_tasks_code ON scheduler_tasks (code);
