-- ============================================================
-- 000057_schedule_role_permissions_least_privilege.up.sql
-- #165 的 000054 将 schedule 四权一并授给 minister / vice_minister（department）。
-- 写入接口按权限码门禁，再按资源 schedule 取最宽 data_scope，
-- 因而副部长可改删同部门他人部门日历，部长可删除该类日历。
-- 与 seed_rbac.go 对齐：部长无 delete，副部长无 update/delete。
-- ============================================================

DELETE FROM role_permissions rp
USING roles r, permissions p
WHERE rp.role_id = r.id
  AND rp.permission_id = p.id
  AND r.code = 'minister'
  AND p.code = 'schedule:delete';

DELETE FROM role_permissions rp
USING roles r, permissions p
WHERE rp.role_id = r.id
  AND rp.permission_id = p.id
  AND r.code = 'vice_minister'
  AND p.code IN ('schedule:update', 'schedule:delete');
