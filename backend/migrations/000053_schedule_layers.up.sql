-- ============================================================
-- 000053_schedule_layers.up.sql
-- 日历图层（personal/timetable/import/google）与导入覆盖字段
-- ============================================================

ALTER TABLE calendars
    ADD COLUMN IF NOT EXISTS source VARCHAR(32) NOT NULL DEFAULT 'personal',
    ADD COLUMN IF NOT EXISTS source_key VARCHAR(200) NOT NULL DEFAULT '';

UPDATE calendars SET source = 'personal' WHERE source IS NULL OR TRIM(source) = '';

DROP INDEX IF EXISTS idx_calendars_personal_owner;
CREATE UNIQUE INDEX IF NOT EXISTS idx_calendars_personal_owner
    ON calendars(owner_id) WHERE calendar_type = 1 AND source = 'personal';

-- 每个用户至多一层课表、一层 Google
CREATE UNIQUE INDEX IF NOT EXISTS idx_calendars_owner_system_layer
    ON calendars(owner_id, source) WHERE source IN ('timetable', 'google');

CREATE INDEX IF NOT EXISTS idx_calendars_source
    ON calendars(owner_id, source, source_key);

ALTER TABLE schedule_events
    ADD COLUMN IF NOT EXISTS origin VARCHAR(32) NOT NULL DEFAULT 'manual',
    ADD COLUMN IF NOT EXISTS external_uid VARCHAR(200) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_schedule_events_origin
    ON schedule_events(calendar_id, origin);

CREATE TABLE IF NOT EXISTS schedule_google_accounts (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    calendar_id   UUID REFERENCES calendars(id) ON DELETE SET NULL,
    access_token  TEXT NOT NULL DEFAULT '',
    refresh_token TEXT NOT NULL DEFAULT '',
    token_expiry  TIMESTAMPTZ,
    google_email  VARCHAR(200) NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO notification_templates
    (id, code, name, title_template, body_template, channels, category, variables_schema, status)
VALUES (
    uuid_generate_v4(),
    'schedule_reminder',
    '日程提醒',
    '日程提醒：{{.title}}',
    '「{{.title}}」将于 {{.start_at}} 开始（提前 {{.minutes}} 分钟）。',
    '["in_app","websocket","email"]',
    'schedule',
    '{"title":"string","start_at":"string","minutes":"string"}'::jsonb,
    0
)
ON CONFLICT (code) DO NOTHING;
