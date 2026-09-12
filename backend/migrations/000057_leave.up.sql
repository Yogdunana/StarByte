-- ============================================================
-- 000057_leave.up.sql
-- Issue #56 phase-1：请假类型 / 余额 / 申请
-- 领域规则来自 SMB-Star PR #162（提交扣余额、驳回返还）
-- 错误码使用 30000-30999
-- ============================================================

CREATE TABLE IF NOT EXISTS leave_types (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(50) NOT NULL,
    code          VARCHAR(20) NOT NULL,
    deductible    BOOLEAN NOT NULL DEFAULT TRUE,
    default_days  NUMERIC(6,1) NOT NULL DEFAULT 0,
    description   VARCHAR(255) NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMPTZ,
    CONSTRAINT uq_leave_types_code UNIQUE (code)
);

COMMENT ON COLUMN leave_types.deductible IS '是否在提交时扣减 / 驳回时返还余额（#162）';
COMMENT ON COLUMN leave_types.default_days IS '首次查看余额时自动开户的默认额度';

CREATE TABLE IF NOT EXISTS leave_balances (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    year           INTEGER NOT NULL,
    leave_type_id  UUID NOT NULL REFERENCES leave_types(id) ON DELETE RESTRICT,
    total_days     NUMERIC(6,1) NOT NULL DEFAULT 0,
    used_days      NUMERIC(6,1) NOT NULL DEFAULT 0,
    remaining_days NUMERIC(6,1) NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at     TIMESTAMPTZ,
    CONSTRAINT uq_leave_balances_user_year_type UNIQUE (user_id, year, leave_type_id)
);

CREATE INDEX IF NOT EXISTS idx_leave_balances_user ON leave_balances(user_id, year);

CREATE TABLE IF NOT EXISTS leave_applications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    applicant_id    UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    leave_type_id   UUID NOT NULL REFERENCES leave_types(id) ON DELETE RESTRICT,
    start_time      TIMESTAMPTZ NOT NULL,
    end_time        TIMESTAMPTZ NOT NULL,
    duration_days   NUMERIC(6,1) NOT NULL,
    reason          VARCHAR(500) NOT NULL DEFAULT '',
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    approver_id     UUID REFERENCES users(id) ON DELETE RESTRICT,
    approve_remark  VARCHAR(500) NOT NULL DEFAULT '',
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMPTZ,
    CONSTRAINT chk_leave_applications_status CHECK (status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT chk_leave_applications_time CHECK (start_time <= end_time)
);

CREATE INDEX IF NOT EXISTS idx_leave_applications_applicant ON leave_applications(applicant_id);
CREATE INDEX IF NOT EXISTS idx_leave_applications_status ON leave_applications(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_leave_applications_type ON leave_applications(leave_type_id);

INSERT INTO leave_types (id, name, code, deductible, default_days, description)
VALUES
    (uuid_generate_v4(), '年假', 'annual', TRUE, 5, '年度带薪假，提交时扣减余额'),
    (uuid_generate_v4(), '调休', 'compensatory', TRUE, 0, '加班调休，提交时扣减余额'),
    (uuid_generate_v4(), '事假', 'personal', FALSE, 0, '事假，不扣减余额'),
    (uuid_generate_v4(), '病假', 'sick', FALSE, 0, '病假，不扣减余额')
ON CONFLICT (code) DO NOTHING;

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '请假查看', 'leave:read', 'leave', 'read', '查看全部请假记录与统计', 3, true, 0),
    (uuid_generate_v4(), '请假审批', 'leave:approve', 'leave', 'approve', '批准或驳回请假申请', 3, true, 0)
ON CONFLICT (code) DO NOTHING;

-- 社长 / 超管 / 副社长：查阅 + 审批
INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'all'
FROM roles r
CROSS JOIN permissions p
WHERE r.code IN ('president', 'super_admin', 'vice_president')
  AND p.code IN ('leave:read', 'leave:approve')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 部长：查阅 + 审批
INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'department'
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'minister'
  AND p.code IN ('leave:read', 'leave:approve')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 副部长 / 干事：查阅全部记录（审批仍由部长及以上处理）
INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'department'
FROM roles r
CROSS JOIN permissions p
WHERE r.code IN ('vice_minister', 'officer')
  AND p.code = 'leave:read'
ON CONFLICT (role_id, permission_id) DO NOTHING;
