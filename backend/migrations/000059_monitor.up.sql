-- ============================================================
-- 000059_monitor.up.sql
-- Issue #87 phase-1：系统监控与运维仪表盘权限
-- 错误码使用 31000-31999
-- ============================================================

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '监控查看', 'monitor:read', 'monitor', 'read', '查看服务器 / 应用 / 数据库 / Redis 运维快照', 3, true, 0)
ON CONFLICT (code) DO NOTHING;

-- 社长 / 超管 / 副社长：运维仪表盘（不授给干事；部长种子里也会排除 monitor）
INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'all'
FROM roles r
CROSS JOIN permissions p
WHERE r.code IN ('president', 'super_admin', 'vice_president')
  AND p.code = 'monitor:read'
ON CONFLICT (role_id, permission_id) DO NOTHING;
