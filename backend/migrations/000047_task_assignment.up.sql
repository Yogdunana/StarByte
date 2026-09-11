ALTER TABLE tasks ADD COLUMN assignment_policy jsonb NOT NULL DEFAULT '{}'::jsonb;
CREATE TABLE task_assignment_cursors (
 id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
 rule_key varchar(160) NOT NULL UNIQUE,
 last_user_id uuid,
 created_at timestamptz NOT NULL DEFAULT NOW(),
 updated_at timestamptz NOT NULL DEFAULT NOW()
);
CREATE INDEX tasks_assignee_live_load ON tasks(assignee_id,status) WHERE deleted_at IS NULL;
