DROP INDEX IF EXISTS idx_feature_exposures_user;
DROP INDEX IF EXISTS idx_feature_exposures_flag_created;
DROP TABLE IF EXISTS feature_flag_exposures;

ALTER TABLE feature_flags DROP CONSTRAINT IF EXISTS chk_feature_flags_type;
ALTER TABLE feature_flags ADD CONSTRAINT chk_feature_flags_type CHECK (
    flag_type IN ('boolean', 'user_allowlist', 'role_dept', 'percentage')
);
