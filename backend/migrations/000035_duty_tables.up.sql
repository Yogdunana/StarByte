-- 000035_duty_tables.up.sql
-- Issue #53: 值班/排班管理

BEGIN;

-- 排班表
CREATE TABLE IF NOT EXISTS duty_schedules (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    duty_date     DATE NOT NULL,
    time_slot     VARCHAR(20) NOT NULL DEFAULT 'full_day', -- morning / afternoon / evening / full_day
    location      VARCHAR(100),
    remark        VARCHAR(500),
    status        SMALLINT NOT NULL DEFAULT 0, -- 0=待值班 1=已到岗 2=已完成 3=缺勤 4=调班
    created_by    UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMP
);

CREATE INDEX idx_duty_schedules_user_id ON duty_schedules(user_id);
CREATE INDEX idx_duty_schedules_department_id ON duty_schedules(department_id);
CREATE INDEX idx_duty_schedules_duty_date ON duty_schedules(duty_date);
CREATE INDEX idx_duty_schedules_status ON duty_schedules(status);
CREATE INDEX idx_duty_schedules_deleted_at ON duty_schedules(deleted_at);
CREATE UNIQUE INDEX idx_duty_schedules_user_date_slot_unique
    ON duty_schedules(user_id, duty_date, time_slot)
    WHERE deleted_at IS NULL;

-- 调班申请表
CREATE TABLE IF NOT EXISTS duty_swap_requests (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    requester_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    requester_schedule_id UUID NOT NULL REFERENCES duty_schedules(id) ON DELETE CASCADE,
    target_schedule_id    UUID REFERENCES duty_schedules(id) ON DELETE CASCADE,
    reason          VARCHAR(500) NOT NULL,
    status          SMALLINT NOT NULL DEFAULT 0, -- 0=待审批 1=已批准 2=已拒绝 3=已取消
    approver_id     UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_at     TIMESTAMP,
    approved_remark VARCHAR(500),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP
);

CREATE INDEX idx_duty_swap_requests_requester_id ON duty_swap_requests(requester_id);
CREATE INDEX idx_duty_swap_requests_target_user_id ON duty_swap_requests(target_user_id);
CREATE INDEX idx_duty_swap_requests_status ON duty_swap_requests(status);
CREATE INDEX idx_duty_swap_requests_deleted_at ON duty_swap_requests(deleted_at);

COMMIT;
