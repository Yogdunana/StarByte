-- ============================================================
-- 000035_activity_tables.up.sql
-- Issue #52：活动管理与报名系统
-- 表：activities, activity_registrations, activity_surveys
-- ============================================================

-- 活动表
CREATE TABLE activities (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title           VARCHAR(200) NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    cover_image_id  UUID REFERENCES files(id) ON DELETE SET NULL,
    category        VARCHAR(50) NOT NULL DEFAULT '',
    tags            JSONB NOT NULL DEFAULT '[]'::jsonb,
    start_time      TIMESTAMP NOT NULL,
    end_time        TIMESTAMP NOT NULL,
    location        VARCHAR(200) NOT NULL DEFAULT '',
    max_participants INTEGER NOT NULL DEFAULT 0,
    status          SMALLINT NOT NULL DEFAULT 0,
    organizer_id    UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at      TIMESTAMP
);

COMMENT ON COLUMN activities.status IS '0=草稿 1=报名中 2=进行中 3=已结束 4=已取消';
COMMENT ON COLUMN activities.max_participants IS '人数上限，0 表示不限';
COMMENT ON COLUMN activities.category IS '活动分类：竞赛/讲座/技术沙龙等';

CREATE INDEX idx_activities_status ON activities(status);
CREATE INDEX idx_activities_category ON activities(category);
CREATE INDEX idx_activities_organizer ON activities(organizer_id);
CREATE INDEX idx_activities_start_time ON activities(start_time);

-- 报名表
CREATE TABLE activity_registrations (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    activity_id      UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    user_id          UUID NOT NULL REFERENCES users(id),
    status           SMALLINT NOT NULL DEFAULT 0,
    checkin_status   SMALLINT NOT NULL DEFAULT 0,
    checked_in_at    TIMESTAMP,
    checkin_method   SMALLINT,
    gps_latitude     DECIMAL(10, 7),
    gps_longitude    DECIMAL(10, 7),
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (activity_id, user_id)
);

COMMENT ON COLUMN activity_registrations.status IS '0=待审批 1=已通过 2=已拒绝 3=候补 4=已取消';
COMMENT ON COLUMN activity_registrations.checkin_status IS '0=未签到 1=已签到';
COMMENT ON COLUMN activity_registrations.checkin_method IS '1=二维码 2=GPS';

CREATE INDEX idx_activity_registrations_activity ON activity_registrations(activity_id);
CREATE INDEX idx_activity_registrations_user ON activity_registrations(user_id);
CREATE INDEX idx_activity_registrations_status ON activity_registrations(status);

-- 满意度调查表
CREATE TABLE activity_surveys (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id),
    rating      SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment     TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (activity_id, user_id)
);

CREATE INDEX idx_activity_surveys_activity ON activity_surveys(activity_id);
CREATE INDEX idx_activity_surveys_user ON activity_surveys(user_id);
