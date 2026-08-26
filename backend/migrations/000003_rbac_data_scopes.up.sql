-- ============================================================
-- 000003_rbac_data_scopes.up.sql
-- RBAC 角色数据权限-自定义部门关联表
-- ============================================================

CREATE TABLE role_data_scopes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_permission_id UUID NOT NULL REFERENCES role_permissions(id) ON DELETE CASCADE,
    department_id UUID NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_permission_id, department_id)
);

CREATE INDEX idx_role_data_scopes_rp_id ON role_data_scopes(role_permission_id);
CREATE INDEX idx_role_data_scopes_dept_id ON role_data_scopes(department_id);
