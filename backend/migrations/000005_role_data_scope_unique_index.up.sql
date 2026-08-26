CREATE UNIQUE INDEX IF NOT EXISTS idx_role_data_scope_role_perm_dept_unique
    ON role_data_scopes (role_permission_id, department_id);
