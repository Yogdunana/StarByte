INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES (uuid_generate_v4(), '统计导出', 'stats:export', 'stats', 'export', '统计导出', 3, true, 0)
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'all'
FROM roles r
CROSS JOIN permissions p
WHERE r.code IN ('president', 'super_admin')
  AND p.code = 'stats:export'
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'department'
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'minister'
  AND p.code = 'stats:export'
ON CONFLICT (role_id, permission_id) DO NOTHING;
