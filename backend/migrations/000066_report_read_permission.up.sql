-- ============================================================
-- 000066_report_read_permission.up.sql
-- Issue #57 phase-1：工作汇报列表读取权限
-- 000064 知识库、000065 备份 phase-2 已占用
-- ============================================================

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES (
    uuid_generate_v4(),
    '工作汇报查看',
    'report:read',
    'report',
    'read',
    '查看工作汇报列表',
    3,
    true,
    0
)
ON CONFLICT (code) DO NOTHING;
