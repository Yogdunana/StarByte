-- 回滚到 000054 对部长/副部长的授权（仅恢复本迁移删除的行）。

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'department'
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'minister'
  AND p.code = 'schedule:delete'
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'department'
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'vice_minister'
  AND p.code IN ('schedule:update', 'schedule:delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;
