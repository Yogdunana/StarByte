-- ============================================================
-- 000067_feature_flags_phase2.up.sql
-- Issue #98 remaining: AB 类型、曝光分析、定时/环境规则已落在 rules JSONB
-- ============================================================

ALTER TABLE feature_flags DROP CONSTRAINT IF EXISTS chk_feature_flags_type;
ALTER TABLE feature_flags ADD CONSTRAINT chk_feature_flags_type CHECK (
    flag_type IN ('boolean', 'user_allowlist', 'role_dept', 'percentage', 'ab_test')
);

CREATE TABLE IF NOT EXISTS feature_flag_exposures (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    flag_id     UUID,
    flag_key    VARCHAR(128) NOT NULL,
    user_id     UUID,
    variant     VARCHAR(64)  NOT NULL DEFAULT '',
    enabled     BOOLEAN      NOT NULL DEFAULT FALSE,
    reason      VARCHAR(64)  NOT NULL DEFAULT '',
    environment VARCHAR(16)  NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feature_exposures_flag_created
    ON feature_flag_exposures (flag_key, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_feature_exposures_user
    ON feature_flag_exposures (user_id, created_at DESC);
