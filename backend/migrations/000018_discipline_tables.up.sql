-- ============================================================
-- 000018_discipline_tables.up.sql
-- 纪律处分与申诉（Issue #19 补充 000011）
-- ============================================================

CREATE TABLE discipline_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    level SMALLINT NOT NULL DEFAULT 1, -- 1=警告 2=严重警告 3=记过
    status SMALLINT NOT NULL DEFAULT 0, -- 0=生效 1=撤销 2=申诉中
    issued_by UUID REFERENCES users(id),
    issued_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_discipline_records_user_id ON discipline_records(user_id);
CREATE INDEX idx_discipline_records_status ON discipline_records(status);
CREATE INDEX idx_discipline_records_created_at ON discipline_records(created_at);

CREATE TABLE discipline_appeals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    record_id UUID NOT NULL REFERENCES discipline_records(id) ON DELETE CASCADE,
    applicant_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reason TEXT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0, -- 0=待审 1=支持 2=驳回
    reviewer_id UUID REFERENCES users(id),
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_discipline_appeals_record_id ON discipline_appeals(record_id);
CREATE INDEX idx_discipline_appeals_status ON discipline_appeals(status);
