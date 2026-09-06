DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE code = 'audit:report');

DELETE FROM permissions WHERE code = 'audit:report';

DROP INDEX IF EXISTS idx_audit_logs_compliance;
DROP INDEX IF EXISTS idx_audit_logs_entity;

ALTER TABLE audit_logs DROP COLUMN IF EXISTS compliance_flags;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS diff_json;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS after_json;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS before_json;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS entity_id;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS entity_type;
