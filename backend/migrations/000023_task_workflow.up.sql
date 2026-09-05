-- ============================================================
-- 000023_task_workflow.up.sql
-- Issue #9：复用 000013 表，不改已有列名
-- tasks.status 对齐 Issue：0待处理 1进行中 2已完成 3已取消 4已挂起
--   （覆盖 000013 注释：2待审核 3已完成 4已取消；表尚无业务数据）
-- task_comments.author_id 对外映射 user_id
-- task_attachments.file_id 复用 files/MinIO，不另存路径列
-- tags 仍为 VARCHAR，存 JSON 数组字符串
-- ============================================================

ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS due_reminded_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS overdue_reminded_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_tasks_parent_id ON tasks(parent_id);
CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);

ALTER TABLE task_comments
    ADD COLUMN IF NOT EXISTS mentions TEXT NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE task_attachments
    ADD COLUMN IF NOT EXISTS uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS task_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    action_type VARCHAR(32) NOT NULL,
    old_value VARCHAR(500) NOT NULL DEFAULT '',
    new_value VARCHAR(500) NOT NULL DEFAULT '',
    operator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_task_logs_task_id ON task_logs(task_id);
CREATE INDEX IF NOT EXISTS idx_task_logs_created_at ON task_logs(created_at);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '任务分配', 'task:assign', 'task', 'assign', '分配任务负责人', 2, true, 0),
    (uuid_generate_v4(), '任务转办', 'task:transfer', 'task', 'transfer', '转办任务', 2, true, 0),
    (uuid_generate_v4(), '任务评论', 'task:comment', 'task', 'comment', '任务评论与提及', 2, true, 0)
ON CONFLICT (code) DO NOTHING;
