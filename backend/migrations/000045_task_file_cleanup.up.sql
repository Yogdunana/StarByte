CREATE TABLE task_file_deletions (
 id uuid PRIMARY KEY,
 attachment_id uuid NOT NULL UNIQUE,
 task_id uuid NOT NULL REFERENCES tasks(id),
 file_id uuid NOT NULL,
 uploaded_by uuid NOT NULL,
 requested_by uuid NOT NULL,
 status varchar(20) NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','completed')),
 attempts integer NOT NULL DEFAULT 0,
 last_error text NOT NULL DEFAULT '',
 next_attempt_at timestamptz NOT NULL DEFAULT now(),
 completed_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_task_file_cleanup_pending ON task_file_deletions(next_attempt_at,id) WHERE status='pending';
