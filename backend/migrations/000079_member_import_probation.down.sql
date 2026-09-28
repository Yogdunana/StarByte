-- 回滚 000079：去掉批量录入预备期的到期时间列。
DROP INDEX IF EXISTS idx_member_profiles_probation_until;

ALTER TABLE member_profiles DROP COLUMN IF EXISTS probation_until;
