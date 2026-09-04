-- ============================================================
-- 000017_finance_tables.up.sql
-- 财务：分类、预算、流水、报销（Issue #19 补充 000010）
-- ============================================================

CREATE TABLE finance_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    direction SMALLINT NOT NULL DEFAULT 1, -- 1=支出 2=收入
    description VARCHAR(255),
    status SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_finance_categories_status ON finance_categories(status);

CREATE TABLE finance_budgets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID NOT NULL REFERENCES finance_categories(id) ON DELETE RESTRICT,
    department_id UUID REFERENCES departments(id) ON DELETE RESTRICT,
    year SMALLINT NOT NULL,
    amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (category_id, department_id, year)
);

CREATE INDEX idx_finance_budgets_year ON finance_budgets(year);

CREATE TABLE finance_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID NOT NULL REFERENCES finance_categories(id) ON DELETE RESTRICT,
    department_id UUID REFERENCES departments(id) ON DELETE RESTRICT,
    amount DECIMAL(12,2) NOT NULL,
    direction SMALLINT NOT NULL DEFAULT 1,
    occurred_at DATE NOT NULL,
    title VARCHAR(200) NOT NULL,
    remark TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_finance_records_occurred_at ON finance_records(occurred_at);
CREATE INDEX idx_finance_records_category_id ON finance_records(category_id);
CREATE INDEX idx_finance_records_department_id ON finance_records(department_id);

CREATE TABLE finance_reimbursements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    applicant_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    category_id UUID REFERENCES finance_categories(id) ON DELETE RESTRICT,
    amount DECIMAL(12,2) NOT NULL,
    title VARCHAR(200) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0, -- 0=待审 1=通过 2=驳回
    reviewer_id UUID REFERENCES users(id),
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_finance_reimbursements_applicant_id ON finance_reimbursements(applicant_id);
CREATE INDEX idx_finance_reimbursements_status ON finance_reimbursements(status);
CREATE INDEX idx_finance_reimbursements_created_at ON finance_reimbursements(created_at);
