-- ============================================================
-- 000004_rbac_unique_indexes.down.sql
-- 回滚 role_permissions 和 user_roles 的联合唯一索引
-- ============================================================

DROP INDEX IF EXISTS idx_role_perms_role_perm_unique;
DROP INDEX IF EXISTS idx_user_roles_user_role_unique;
