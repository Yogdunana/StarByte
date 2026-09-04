DELETE FROM permissions WHERE code IN ('audit:read', 'audit:export', 'audit:archive');

DROP TABLE IF EXISTS audit_log_archives;

DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_module;

ALTER TABLE audit_logs DROP COLUMN IF EXISTS real_name;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS action;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS module;
