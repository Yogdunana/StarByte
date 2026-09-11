-- One activation identifies the approvals created by one entry into a node.
-- Historical tasks remain NULL; never infer previous signatures or generations.
ALTER TABLE flow_tasks ADD COLUMN IF NOT EXISTS activation_id UUID;
CREATE INDEX IF NOT EXISTS idx_flow_tasks_activation ON flow_tasks (activation_id, status);
