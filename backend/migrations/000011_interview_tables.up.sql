-- ============================================================
-- 000011_interview_tables.up.sql
-- 面试安排、面试官关联、评分（Issue #19 原 000004）
-- ============================================================

CREATE TABLE interviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES member_applications(id) ON DELETE CASCADE,
    round SMALLINT NOT NULL DEFAULT 1,
    type SMALLINT NOT NULL DEFAULT 1, -- 1=技术面 2=综合面 3=HR面
    status SMALLINT NOT NULL DEFAULT 0, -- 0=待安排 1=已安排 2=进行中 3=已完成
    scheduled_at TIMESTAMP,
    location VARCHAR(200),
    duration INT, -- 分钟
    score DECIMAL(5,2),
    result VARCHAR(50),
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_interviews_application_id ON interviews(application_id);
CREATE INDEX idx_interviews_status ON interviews(status);
CREATE INDEX idx_interviews_scheduled_at ON interviews(scheduled_at);
CREATE INDEX idx_interviews_created_at ON interviews(created_at);

CREATE TABLE interview_interviewers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    interview_id UUID NOT NULL REFERENCES interviews(id) ON DELETE CASCADE,
    interviewer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (interview_id, interviewer_id)
);

CREATE INDEX idx_interview_interviewers_interviewer_id ON interview_interviewers(interviewer_id);

CREATE TABLE interview_evaluations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    interview_id UUID NOT NULL REFERENCES interviews(id) ON DELETE CASCADE,
    interviewer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    score DECIMAL(5,2) NOT NULL,
    comment TEXT,
    recommendation SMALLINT NOT NULL DEFAULT 3, -- 1=强烈推荐 2=推荐 3=待定 4=不推荐
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (interview_id, interviewer_id)
);

CREATE INDEX idx_interview_evaluations_interview_id ON interview_evaluations(interview_id);
