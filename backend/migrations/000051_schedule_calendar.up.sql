-- ============================================================
-- 000051_schedule_calendar.up.sql
-- Issue #78：个人/共享日历与日程事件（PostgreSQL，非 000026 调度引擎）
-- ============================================================

CREATE TABLE IF NOT EXISTS calendars (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(200) NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    calendar_type SMALLINT NOT NULL DEFAULT 1, -- 1=个人 2=部门 3=项目/共享
    color         VARCHAR(16) NOT NULL DEFAULT '#2563eb',
    owner_id      UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_calendars_owner ON calendars(owner_id);
CREATE INDEX IF NOT EXISTS idx_calendars_dept ON calendars(department_id);
CREATE INDEX IF NOT EXISTS idx_calendars_type ON calendars(calendar_type);
CREATE UNIQUE INDEX IF NOT EXISTS idx_calendars_personal_owner
    ON calendars(owner_id) WHERE calendar_type = 1;

CREATE TABLE IF NOT EXISTS calendar_members (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calendar_id UUID NOT NULL REFERENCES calendars(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        SMALLINT NOT NULL DEFAULT 1, -- 1=查看 2=编辑
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (calendar_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_calendar_members_user ON calendar_members(user_id);

CREATE TABLE IF NOT EXISTS schedule_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calendar_id     UUID NOT NULL REFERENCES calendars(id) ON DELETE CASCADE,
    title           VARCHAR(200) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    location        VARCHAR(200) NOT NULL DEFAULT '',
    start_at        TIMESTAMPTZ NOT NULL,
    end_at          TIMESTAMPTZ NOT NULL,
    all_day         BOOLEAN NOT NULL DEFAULT FALSE,
    color           VARCHAR(16) NOT NULL DEFAULT '',
    status          SMALLINT NOT NULL DEFAULT 0, -- 0=确认 1=取消
    recurrence      VARCHAR(16) NOT NULL DEFAULT 'none', -- none/daily/weekly/monthly
    recurrence_until TIMESTAMPTZ,
    meeting_id      UUID, -- 软关联会议，无 FK
    created_by      UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_schedule_events_time CHECK (end_at >= start_at)
);

CREATE INDEX IF NOT EXISTS idx_schedule_events_calendar ON schedule_events(calendar_id, start_at);
CREATE INDEX IF NOT EXISTS idx_schedule_events_start ON schedule_events(start_at, end_at);
CREATE INDEX IF NOT EXISTS idx_schedule_events_meeting ON schedule_events(meeting_id);
CREATE INDEX IF NOT EXISTS idx_schedule_events_creator ON schedule_events(created_by);

CREATE TABLE IF NOT EXISTS schedule_event_attendees (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id        UUID NOT NULL REFERENCES schedule_events(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    response_status SMALLINT NOT NULL DEFAULT 0, -- 0=待回复 1=接受 2=拒绝 3=待定
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (event_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_schedule_attendees_user ON schedule_event_attendees(user_id);

CREATE TABLE IF NOT EXISTS schedule_reminders (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id       UUID NOT NULL REFERENCES schedule_events(id) ON DELETE CASCADE,
    minutes_before INT NOT NULL,
    method         SMALLINT NOT NULL DEFAULT 1, -- 1=站内
    triggered_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (event_id, minutes_before, method),
    CONSTRAINT chk_schedule_reminder_minutes CHECK (minutes_before IN (5, 15, 30, 60))
);

CREATE INDEX IF NOT EXISTS idx_schedule_reminders_pending
    ON schedule_reminders(triggered_at, minutes_before);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '日程查看', 'schedule:read', 'schedule', 'read', '查看日历与事件', 3, true, 0),
    (uuid_generate_v4(), '日程创建', 'schedule:create', 'schedule', 'create', '创建日历与事件', 3, true, 0),
    (uuid_generate_v4(), '日程更新', 'schedule:update', 'schedule', 'update', '更新日历与事件', 3, true, 0),
    (uuid_generate_v4(), '日程删除', 'schedule:delete', 'schedule', 'delete', '删除日历与事件', 3, true, 0)
ON CONFLICT (code) DO NOTHING;

INSERT INTO scheduler_tasks (name, code, cron_expr, timezone, handler_key, next_run_at)
SELECT '日程提醒扫描', 'schedule_reminder', '0 * * * * *', 'Asia/Shanghai', 'schedule_reminder', NOW()
WHERE NOT EXISTS (SELECT 1 FROM scheduler_tasks WHERE code = 'schedule_reminder' AND status <> 2);
