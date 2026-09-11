-- ============================================================
-- 000052_user_identity_type_value_unique.up.sql
-- 同一外部身份（CAS / 学号等）只能绑定一个本地用户
-- ============================================================

CREATE UNIQUE INDEX IF NOT EXISTS uk_user_identities_type_value
    ON user_identities (identity_type, identity_value);
