DO $$ BEGIN
 IF EXISTS (SELECT 1 FROM flow_tasks WHERE activation_id IS NOT NULL) THEN
  RAISE EXCEPTION 'Approval activation history exists; retain it or export and migrate explicitly';
 END IF;
END $$;
DROP INDEX IF EXISTS idx_flow_tasks_activation;
ALTER TABLE flow_tasks DROP COLUMN IF EXISTS activation_id;
