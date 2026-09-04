-- ============================================================
-- 000014_internship_tables.up.sql
-- 实习时段 + 打卡记录（Issue #19 原 000007）
-- ============================================================

CREATE TABLE internships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    department_id UUID REFERENCES departments(id) ON DELETE RESTRICT,
    start_date DATE NOT NULL,
    end_date DATE,
    status SMALLINT NOT NULL DEFAULT 0, -- 0=进行中 1=已结束 2=已取消
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_internships_user_id ON internships(user_id);
CREATE INDEX idx_internships_department_id ON internships(department_id);
CREATE INDEX idx_internships_status ON internships(status);

CREATE TABLE internship_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    internship_id UUID REFERENCES internships(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    department_id UUID REFERENCES departments(id) ON DELETE RESTRICT,
    work_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    duration DECIMAL(6,2) NOT NULL DEFAULT 0, -- 小时
    task_description TEXT,
    status SMALLINT NOT NULL DEFAULT 0, -- 0=待审核 1=已通过 2=已拒绝
    reviewer_id UUID REFERENCES users(id),
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_internship_records_user_id ON internship_records(user_id);
CREATE INDEX idx_internship_records_work_date ON internship_records(work_date);
CREATE INDEX idx_internship_records_status ON internship_records(status);
CREATE INDEX idx_internship_records_department_id ON internship_records(department_id);
CREATE INDEX idx_internship_records_created_at ON internship_records(created_at);
