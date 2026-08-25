-- ============================================================
-- 000004_rbac_unique_indexes.up.sql
-- 为 role_permissions 和 user_roles 添加联合唯一索引
-- ============================================================

-- role_permissions: 同一角色下同一权限只允许关联一次
CREATE UNIQUE INDEX IF NOT EXISTS idx_role_perms_role_perm_unique
    ON role_permissions (role_id, permission_id);

-- user_roles: 同一用户下同一角色只允许关联一次
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_user_role_unique
    ON user_roles (user_id, role_id);
