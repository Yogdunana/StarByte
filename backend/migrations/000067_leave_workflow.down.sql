-- 000067 only publishes leave_approval when none exists.
-- Rolling back must not drop a designer-published graph.
-- Column drops are reversible for local/dev rollback only.

ALTER TABLE leave_applications
    DROP COLUMN IF EXISTS attachments,
    DROP COLUMN IF EXISTS workflow_stage,
    DROP COLUMN IF EXISTS workflow_instance_id;

ALTER TABLE leave_types
    DROP COLUMN IF EXISTS sort_order,
    DROP COLUMN IF EXISTS enabled;
