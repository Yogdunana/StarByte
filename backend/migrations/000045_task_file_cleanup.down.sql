DO $$ BEGIN IF EXISTS(SELECT 1 FROM task_file_deletions WHERE status='pending') THEN RAISE EXCEPTION 'Complete pending file cleanup before downgrade'; END IF; END $$;
DROP TABLE task_file_deletions;
