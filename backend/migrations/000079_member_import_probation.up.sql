-- 批量录入的预备干事需要一个到期时间，否则没人知道什么时候该转正式干事。
--
-- 入会流程的预备期记在 member_applications.probation_until 上，而批量录入的人没有
-- 对应的申请记录，所以档案自己也要有一列。
ALTER TABLE member_profiles ADD COLUMN IF NOT EXISTS probation_until TIMESTAMPTZ;

COMMENT ON COLUMN member_profiles.probation_until IS
  '预备期届满时间，仅 status=3（预备期）时有意义；NULL 表示不在预备期内';

-- 维护任务每分钟扫一次到期的人，这里给个小索引就够了。
CREATE INDEX IF NOT EXISTS idx_member_profiles_probation_until
  ON member_profiles (probation_until)
  WHERE status = 3 AND probation_until IS NOT NULL;
