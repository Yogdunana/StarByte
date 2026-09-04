-- ============================================================
-- 000021_interview_management.up.sql
-- Issue #7：场次、维度、补齐 interviews / evaluations 列与权限
-- 复用 000011 表，不改已有列名（scheduled_at / interviewer_id）
-- interviews.status 对齐 Issue：0待面试 1已签到 2面试中 3已完成 4缺席 5已取消
-- ============================================================

CREATE TABLE IF NOT EXISTS interview_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    round SMALLINT NOT NULL DEFAULT 1,
    department_id UUID REFERENCES departments(id) ON DELETE RESTRICT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    location VARCHAR(200) NOT NULL DEFAULT '',
    online_link VARCHAR(500) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 0, -- 0待开始 1进行中 2已结束 3已取消
    max_candidates INT NOT NULL DEFAULT 20,
    description TEXT NOT NULL DEFAULT '',
    created_by UUID REFERENCES users(id),
    qr_token VARCHAR(64) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_interview_sessions_status ON interview_sessions(status);
CREATE INDEX IF NOT EXISTS idx_interview_sessions_department ON interview_sessions(department_id);
CREATE INDEX IF NOT EXISTS idx_interview_sessions_round ON interview_sessions(round);

CREATE TABLE IF NOT EXISTS interview_dimensions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL,
    weight NUMERIC(5,2) NOT NULL DEFAULT 1,
    max_score NUMERIC(5,2) NOT NULL DEFAULT 100,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_interview_dimensions_name ON interview_dimensions(name);

ALTER TABLE interviews
    ALTER COLUMN application_id DROP NOT NULL,
    ADD COLUMN IF NOT EXISTS session_id UUID REFERENCES interview_sessions(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS applicant_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS actual_start_time TIMESTAMP,
    ADD COLUMN IF NOT EXISTS actual_end_time TIMESTAMP,
    ADD COLUMN IF NOT EXISTS result_code SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS result_comment TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_interviews_session_id ON interviews(session_id);
CREATE INDEX IF NOT EXISTS idx_interviews_applicant_id ON interviews(applicant_id);

ALTER TABLE interview_evaluations
    ADD COLUMN IF NOT EXISTS dimension VARCHAR(50) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

DROP INDEX IF EXISTS interview_evaluations_interview_id_interviewer_id_key;
CREATE UNIQUE INDEX IF NOT EXISTS uq_interview_eval_dim
    ON interview_evaluations (interview_id, interviewer_id, dimension);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '面试管理', 'interview:manage', 'interview', 'manage', '场次与结果管理', 2, true, 0),
    (uuid_generate_v4(), '面试评分', 'interview:evaluate', 'interview', 'evaluate', '面试官评分与开场', 2, true, 0)
ON CONFLICT (code) DO NOTHING;

INSERT INTO interview_dimensions (id, name, weight, max_score, sort_order)
VALUES
    (uuid_generate_v4(), '技术能力', 0.40, 100, 1),
    (uuid_generate_v4(), '沟通能力', 0.30, 100, 2),
    (uuid_generate_v4(), '逻辑思维', 0.30, 100, 3)
ON CONFLICT (name) DO NOTHING;
