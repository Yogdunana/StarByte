-- ============================================================
-- 000060_backup.up.sql
-- Issue #88 phase-1：数据备份与恢复（全量 pg_dump + gzip + MinIO）
-- 错误码使用 32000-32999
-- 权限仅授给 president / super_admin / vice_president（与 monitor 同运维口径）
-- ============================================================

CREATE TABLE IF NOT EXISTS backup_records (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    trigger_source  VARCHAR(16)  NOT NULL DEFAULT 'manual',
    status          SMALLINT     NOT NULL DEFAULT 0,
    storage         VARCHAR(16)  NOT NULL DEFAULT '',
    object_key      TEXT         NOT NULL DEFAULT '',
    filename        VARCHAR(255) NOT NULL DEFAULT '',
    checksum_sha256 VARCHAR(64)  NOT NULL DEFAULT '',
    size_bytes      BIGINT       NOT NULL DEFAULT 0,
    started_at      TIMESTAMPTZ,
    finished_at     TIMESTAMPTZ,
    error_message   TEXT         NOT NULL DEFAULT '',
    created_by      UUID,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_backup_records_status_created
    ON backup_records (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_backup_records_finished
    ON backup_records (finished_at);

CREATE TABLE IF NOT EXISTS backup_policies (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    enabled        BOOLEAN      NOT NULL DEFAULT FALSE,
    retention_days INT          NOT NULL DEFAULT 30,
    cron_expr      VARCHAR(64)  NOT NULL DEFAULT '0 30 2 * * *',
    timezone       VARCHAR(64)  NOT NULL DEFAULT 'Asia/Shanghai',
    updated_by     UUID,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_backup_policies_retention CHECK (retention_days >= 1 AND retention_days <= 3650)
);

INSERT INTO backup_policies (id, enabled, retention_days, cron_expr, timezone)
SELECT '00000000-0000-4000-8000-000000000060', FALSE, 30, '0 30 2 * * *', 'Asia/Shanghai'
WHERE NOT EXISTS (SELECT 1 FROM backup_policies);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '备份查看', 'backup:read', 'backup', 'read', '查看备份记录、策略与存储统计', 3, true, 0),
    (uuid_generate_v4(), '备份创建', 'backup:create', 'backup', 'create', '手动触发全量备份', 3, true, 0),
    (uuid_generate_v4(), '备份删除', 'backup:delete', 'backup', 'delete', '删除备份产物与记录', 3, true, 0),
    (uuid_generate_v4(), '备份恢复', 'backup:restore', 'backup', 'restore', '从备份恢复数据库（危险操作）', 3, true, 0),
    (uuid_generate_v4(), '备份管理', 'backup:manage', 'backup', 'manage', '修改备份保留与调度策略', 3, true, 0)
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'all'
FROM roles r
CROSS JOIN permissions p
WHERE r.code IN ('president', 'super_admin', 'vice_president')
  AND p.code IN ('backup:read', 'backup:create', 'backup:delete', 'backup:restore', 'backup:manage')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO notification_templates
    (id, code, name, title_template, body_template, channels, category, variables_schema, status)
VALUES (
    uuid_generate_v4(),
    'backup_failed',
    '备份失败',
    '备份失败：{{.filename}}',
    '备份任务失败：{{.error}}',
    '["in_app","websocket"]',
    'system',
    '{"filename":"string","error":"string"}'::jsonb,
    0
)
ON CONFLICT (code) DO NOTHING;

-- 默认为暂停；策略启用后由服务同步 cron / status
INSERT INTO scheduler_tasks (name, code, cron_expr, timezone, handler_key, next_run_at, timeout_sec, max_retries, status)
SELECT '数据库全量备份', 'backup_scheduled_full', '0 30 2 * * *', 'Asia/Shanghai', 'backup_scheduled_full', NOW(), 1800, 1, 1
WHERE NOT EXISTS (SELECT 1 FROM scheduler_tasks WHERE code = 'backup_scheduled_full' AND status <> 2);

INSERT INTO scheduler_tasks (name, code, cron_expr, timezone, handler_key, next_run_at, timeout_sec, max_retries, status)
SELECT '备份保留清理', 'backup_retention_cleanup', '0 0 4 * * *', 'Asia/Shanghai', 'backup_retention_cleanup', NOW(), 300, 1, 0
WHERE NOT EXISTS (SELECT 1 FROM scheduler_tasks WHERE code = 'backup_retention_cleanup' AND status <> 2);
