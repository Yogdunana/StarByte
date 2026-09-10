CREATE INDEX IF NOT EXISTS idx_meeting_votes_expiry ON meeting_votes(end_time,id) WHERE status=1;
INSERT INTO scheduler_tasks (name,code,cron_expr,timezone,handler_key,next_run_at)
SELECT '会议投票到期关闭','meeting_vote_expiry','0 * * * * *','Asia/Shanghai','meeting_vote_expiry',NOW()
WHERE NOT EXISTS (SELECT 1 FROM scheduler_tasks WHERE code='meeting_vote_expiry' AND status<>2);
