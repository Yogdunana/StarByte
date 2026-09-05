-- ============================================================
-- 000022_meeting_management.up.sql
-- Issue #8：复用 000012 表，不改已有列名
-- meetings.status 对齐 Issue：0待开始 1进行中 2已结束 3已取消
-- meetings.meeting_type 保持 000012：1例会 2临时 3线上（投票类型在 meeting_votes）
-- meeting_agendas.description 对外映射 content；speaker_id 保留
-- meeting_vote_records.voter_id 对外映射 user_id
-- ============================================================

ALTER TABLE meetings
    ADD COLUMN IF NOT EXISTS online_link VARCHAR(500) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS minutes TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS qr_token VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS cancel_reason VARCHAR(500) NOT NULL DEFAULT '';

ALTER TABLE meeting_agendas
    ADD COLUMN IF NOT EXISTS presenter VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE meeting_attendees
    ADD COLUMN IF NOT EXISTS checked_in_at TIMESTAMP;

ALTER TABLE meeting_votes
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE meeting_vote_options
    ADD COLUMN IF NOT EXISTS option_key VARCHAR(64) NOT NULL DEFAULT '';

UPDATE meeting_vote_options
SET option_key = 'opt_' || id::text
WHERE option_key = '' OR option_key IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_meeting_vote_options_key
    ON meeting_vote_options (vote_id, option_key);

ALTER TABLE meeting_vote_records
    ADD COLUMN IF NOT EXISTS weight NUMERIC(8,2) NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS option_key VARCHAR(64) NOT NULL DEFAULT '';

UPDATE meeting_vote_records r
SET option_key = o.option_key
FROM meeting_vote_options o
WHERE r.option_id = o.id AND (r.option_key = '' OR r.option_key IS NULL);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '会议管理', 'meeting:manage', 'meeting', 'manage', '会议状态、议程、参会人与投票管理', 2, true, 0)
ON CONFLICT (code) DO NOTHING;

INSERT INTO configs (id, config_key, config_value, config_type, description, category, is_public)
VALUES (
    uuid_generate_v4(),
    'vote_weight_config',
    '{"weights":{"president":5,"vice_president":4,"minister":3,"deputy":2,"vice_minister":2,"officer":1},"default_weight":1}',
    'json',
    '会议加权投票职务权重',
    'meeting',
    false
)
ON CONFLICT (config_key) DO NOTHING;
