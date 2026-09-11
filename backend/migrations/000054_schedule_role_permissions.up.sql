-- ============================================================
-- 000054_schedule_role_permissions.up.sql
-- 000051 仅插入 schedule:* 权限，未写入 role_permissions。
-- 生产已 migrate 51/53 但未重跑 seed 时，已登录用户访问 /schedule 会 403。
-- 产品意图：每位登录用户可使用个人日程；不放开为无鉴权接口。
-- ============================================================

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '日程查看', 'schedule:read', 'schedule', 'read', '查看日历与事件', 3, true, 0),
    (uuid_generate_v4(), '日程创建', 'schedule:create', 'schedule', 'create', '创建日历与事件', 3, true, 0),
    (uuid_generate_v4(), '日程更新', 'schedule:update', 'schedule', 'update', '更新日历与事件', 3, true, 0),
    (uuid_generate_v4(), '日程删除', 'schedule:delete', 'schedule', 'delete', '删除日历与事件', 3, true, 0)
ON CONFLICT (code) DO NOTHING;

-- 社长 / 超管 / 副社长：全部范围（与 seed_rbac vice_president 对 schedule 的授权一致）
INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'all'
FROM roles r
CROSS JOIN permissions p
WHERE r.code IN ('president', 'super_admin', 'vice_president')
  AND p.code IN ('schedule:read', 'schedule:create', 'schedule:update', 'schedule:delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 部长 / 副部长 / 干事：部门范围（与 seed_rbac data_scope 一致）
INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'department'
FROM roles r
CROSS JOIN permissions p
WHERE r.code IN ('officer', 'minister', 'vice_minister')
  AND p.code IN ('schedule:read', 'schedule:create', 'schedule:update', 'schedule:delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 会员：仅本人
INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'self'
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'member'
  AND p.code IN ('schedule:read', 'schedule:create', 'schedule:update', 'schedule:delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;
