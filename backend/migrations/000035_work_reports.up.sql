-- Work reports for Issue #57.

CREATE TABLE work_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    report_type VARCHAR(16) NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    work_content TEXT NOT NULL DEFAULT '',
    next_plan TEXT NOT NULL DEFAULT '',
    problems_suggestions TEXT NOT NULL DEFAULT '',
    review_status VARCHAR(16) NOT NULL DEFAULT 'pending',
    submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_work_reports_type
        CHECK (report_type IN ('daily', 'weekly', 'monthly')),
    CONSTRAINT chk_work_reports_review_status
        CHECK (review_status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT chk_work_reports_period
        CHECK (period_end >= period_start),
    CONSTRAINT uq_work_reports_user_type_period
        UNIQUE (user_id, report_type, period_start, period_end)
);

CREATE INDEX idx_work_reports_user_period
    ON work_reports(user_id, period_start DESC, period_end DESC);
CREATE INDEX idx_work_reports_department_period
    ON work_reports(department_id, period_start DESC, period_end DESC);
CREATE INDEX idx_work_reports_review_status
    ON work_reports(review_status);
