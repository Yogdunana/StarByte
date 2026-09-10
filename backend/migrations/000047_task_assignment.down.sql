DO $$ BEGIN IF EXISTS(SELECT 1 FROM tasks WHERE assignment_policy<>'{}'::jsonb) THEN RAISE EXCEPTION 'Automatic assignment audit exists; use a forward migration'; END IF; END $$;
DROP INDEX tasks_assignee_live_load;
DROP TABLE task_assignment_cursors;
ALTER TABLE tasks DROP COLUMN assignment_policy;
