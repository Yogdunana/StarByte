-- Refuse a downgrade that would silently resurrect deleted tasks.
DO $$ BEGIN
 IF EXISTS (SELECT 1 FROM tasks WHERE deleted_at IS NOT NULL) THEN
  RAISE EXCEPTION 'Cannot remove task deletion history: restore or archive deleted tasks explicitly first';
 END IF;
END $$;
DROP INDEX IF EXISTS idx_tasks_live_department;
DROP INDEX IF EXISTS idx_tasks_live_assignee;
ALTER TABLE tasks DROP COLUMN deleted_at;
