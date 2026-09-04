-- ============================================================
-- 000006_audit_log_system.up.sql
-- 审计日志：模块/动作/姓名字段 + 归档表 + 权限码
-- ============================================================

ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS module VARCHAR(50);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS action VARCHAR(20);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS real_name VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_audit_logs_module ON audit_logs(module);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);

CREATE TABLE IF NOT EXISTS audit_log_archives (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    archive_date VARCHAR(10) NOT NULL,
    record_count BIGINT DEFAULT 0,
    minio_object VARCHAR(500),
    status SMALLINT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_log_archives_date ON audit_log_archives(archive_date);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '审计日志查看', 'audit:read', 'audit', 'read', '查询审计日志列表与详情', 3, true, 0),
    (uuid_generate_v4(), '审计日志导出', 'audit:export', 'audit', 'export', '导出审计日志 CSV/Excel', 2, true, 0),
    (uuid_generate_v4(), '审计日志归档', 'audit:archive', 'audit', 'archive', '手动触发审计日志归档', 2, true, 0)
ON CONFLICT (code) DO NOTHING;
