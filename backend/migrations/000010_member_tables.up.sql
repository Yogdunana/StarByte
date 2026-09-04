-- ============================================================
-- 000010_member_tables.up.sql
-- 会员：入会申请 + 档案；补齐 users 对部门/职位的外键（删除时 RESTRICT）
-- Issue #19 原计划 000003，因序号已被占用后移
-- ============================================================

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_users_department'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT fk_users_department
            FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE RESTRICT;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_users_position'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT fk_users_position
            FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE RESTRICT;
    END IF;
END $$;

CREATE TABLE member_applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type SMALLINT NOT NULL DEFAULT 1, -- 1=会员 2=干事
    department_id UUID REFERENCES departments(id) ON DELETE RESTRICT,
    reason TEXT,
    status SMALLINT NOT NULL DEFAULT 0, -- 0=待审核 1=一面中 2=二面中 3=已通过 4=已拒绝 5=已取消
    current_stage VARCHAR(50),
    submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reviewed_at TIMESTAMP,
    reviewer_id UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_member_applications_user_id ON member_applications(user_id);
CREATE INDEX idx_member_applications_status ON member_applications(status);
CREATE INDEX idx_member_applications_department_id ON member_applications(department_id);
CREATE INDEX idx_member_applications_created_at ON member_applications(created_at);

CREATE TABLE member_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    member_type SMALLINT NOT NULL DEFAULT 1, -- 1=会员 2=干事
    department_id UUID REFERENCES departments(id) ON DELETE RESTRICT,
    position_id UUID REFERENCES positions(id) ON DELETE RESTRICT,
    join_date DATE,
    status SMALLINT NOT NULL DEFAULT 0, -- 0=正常 1=停用
    points INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_member_profiles_department_id ON member_profiles(department_id);
CREATE INDEX idx_member_profiles_status ON member_profiles(status);
CREATE INDEX idx_member_profiles_member_type ON member_profiles(member_type);
