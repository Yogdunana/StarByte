-- Keep execution logs and pause the task on rollback.
UPDATE scheduler_tasks SET status=1,updated_at=NOW() WHERE code='admission_maintenance' AND status<>2;
