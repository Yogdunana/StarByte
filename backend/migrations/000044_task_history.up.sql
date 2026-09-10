-- Hide deleted tasks while retaining comments, logs and attachment references.
ALTER TABLE tasks ADD COLUMN deleted_at timestamptz;
CREATE INDEX idx_tasks_live_assignee ON tasks(assignee_id,status,due_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_live_department ON tasks(department_id,created_at DESC) WHERE deleted_at IS NULL;
