-- Issue #28：动态表单引擎。#28 原文错误码 16001-16099 已被会话占用，业务码用 22000 段。

CREATE TABLE IF NOT EXISTS forms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 0,
    fields JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_forms_status ON forms(status);
CREATE INDEX IF NOT EXISTS idx_forms_deleted_at ON forms(deleted_at);
CREATE INDEX IF NOT EXISTS idx_forms_name ON forms(name);

CREATE TABLE IF NOT EXISTS form_submissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    form_id UUID NOT NULL REFERENCES forms(id),
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    submitted_by UUID,
    submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_form_submissions_form_id ON form_submissions(form_id);
CREATE INDEX IF NOT EXISTS idx_form_submissions_submitted_at ON form_submissions(submitted_at DESC);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '表单查看', 'form:read', 'form', 'read', '查看表单定义与提交记录', 1, true, 0),
    (uuid_generate_v4(), '表单编辑', 'form:write', 'form', 'write', '创建和更新表单定义', 2, true, 0),
    (uuid_generate_v4(), '表单提交', 'form:submit', 'form', 'submit', '填写并提交已发布表单', 2, true, 0)
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'all'
FROM roles r
CROSS JOIN permissions p
WHERE r.code IN ('president', 'super_admin')
  AND p.code IN ('form:read', 'form:write', 'form:submit')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'department'
FROM roles r
CROSS JOIN permissions p
WHERE r.code IN ('minister', 'vice_minister')
  AND p.code IN ('form:read', 'form:write', 'form:submit')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'department'
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'officer'
  AND p.code IN ('form:read', 'form:submit')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'self'
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'member'
  AND p.code = 'form:submit'
ON CONFLICT (role_id, permission_id) DO NOTHING;
