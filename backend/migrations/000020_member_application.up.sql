-- ============================================================
-- 000020_member_application.up.sql
-- Issue #6：补齐入会申请/档案列、历史表、审核与导出权限
-- 复用 000010 已建表，不改已有列名（申请类型列仍为 type）
-- status 语义对齐 Issue：0待审核 1审核中 2面试中 3通过 4拒绝 5补充材料
-- ============================================================

ALTER TABLE member_applications
    ADD COLUMN IF NOT EXISTS real_name VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS student_no VARCHAR(30) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS skills JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS experience TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS contact_phone VARCHAR(20) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS contact_email VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS flow_instance_id UUID,
    ADD COLUMN IF NOT EXISTS review_comment TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS required_fields JSONB NOT NULL DEFAULT '[]'::jsonb;

CREATE INDEX IF NOT EXISTS idx_member_applications_student_no
    ON member_applications(student_no);
CREATE INDEX IF NOT EXISTS idx_member_applications_type
    ON member_applications(type);

ALTER TABLE member_profiles
    ADD COLUMN IF NOT EXISTS real_name VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS student_no VARCHAR(30) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS gender SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS grade VARCHAR(20) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS major VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS leave_date DATE,
    ADD COLUMN IF NOT EXISTS skills JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS projects JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS bio TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS contact_phone VARCHAR(20) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS contact_email VARCHAR(100) NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS uq_member_profiles_student_no
    ON member_profiles(student_no)
    WHERE student_no <> '';

CREATE TABLE IF NOT EXISTS member_application_histories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES member_applications(id) ON DELETE CASCADE,
    from_status SMALLINT NOT NULL,
    to_status SMALLINT NOT NULL,
    operator_id UUID REFERENCES users(id),
    comment TEXT NOT NULL DEFAULT '',
    extra JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_member_app_hist_app_id
    ON member_application_histories(application_id);
CREATE INDEX IF NOT EXISTS idx_member_app_hist_created
    ON member_application_histories(created_at);

CREATE TABLE IF NOT EXISTS member_profile_histories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    profile_id UUID NOT NULL REFERENCES member_profiles(id) ON DELETE CASCADE,
    field_name VARCHAR(50) NOT NULL,
    old_value TEXT NOT NULL DEFAULT '',
    new_value TEXT NOT NULL DEFAULT '',
    operator_id UUID REFERENCES users(id),
    reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_member_prof_hist_profile_id
    ON member_profile_histories(profile_id);
CREATE INDEX IF NOT EXISTS idx_member_prof_hist_created
    ON member_profile_histories(created_at);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '会员审核', 'member:approve', 'member', 'approve', '审核入会申请', 2, true, 0),
    (uuid_generate_v4(), '档案导出', 'member:export', 'member', 'export', '导出人员档案 PDF', 3, true, 0),
    (uuid_generate_v4(), '会员管理', 'member:manage', 'member', 'manage', '变更会员档案状态', 2, true, 0)
ON CONFLICT (code) DO NOTHING;
