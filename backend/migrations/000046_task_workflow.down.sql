DO $migration$
BEGIN
 IF EXISTS(SELECT 1 FROM tasks WHERE workflow_instance_id IS NOT NULL) THEN
  RAISE EXCEPTION 'Cannot remove task workflow history; use a forward migration';
 END IF;
END $migration$;
ALTER TABLE tasks DROP COLUMN workflow_revision, DROP COLUMN submission, DROP COLUMN acceptor_id, DROP COLUMN reviewer_id, DROP COLUMN workflow_stage, DROP COLUMN workflow_instance_id;
-- Keep published definition versions for audit; no destructive template deletion.
