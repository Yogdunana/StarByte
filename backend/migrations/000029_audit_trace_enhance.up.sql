-- 000029_audit_trace_enhance.up.sql
-- Issue #76：#5 二期，扩展 audit_logs 列（不新建第二张日志表）

ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS entity_type VARCHAR(50);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS entity_id VARCHAR(64);
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS before_json TEXT;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS after_json TEXT;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS diff_json TEXT;
ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS compliance_flags VARCHAR(200);

CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_compliance ON audit_logs(compliance_flags);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '审计合规报告', 'audit:report', 'audit', 'report', '生成审计合规报告 JSON/CSV/PDF', 2, true, 0)
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'all'
FROM roles r
CROSS JOIN permissions p
WHERE r.code IN ('president', 'super_admin')
  AND p.code = 'audit:report'
ON CONFLICT (role_id, permission_id) DO NOTHING;
