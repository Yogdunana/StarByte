UPDATE scheduler_tasks SET status=1 WHERE code='meeting_vote_expiry' AND status<>2;
DROP INDEX IF EXISTS idx_meeting_votes_expiry;
