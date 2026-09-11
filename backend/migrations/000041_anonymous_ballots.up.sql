-- Separate the participation receipt from the ballot. Preserve counts/choices,
-- but remove identity and exact cast timestamps from historical anonymous ballots.
CREATE TABLE meeting_vote_receipts (
 vote_id UUID NOT NULL REFERENCES meeting_votes(id) ON DELETE CASCADE,
 voter_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
 PRIMARY KEY(vote_id,voter_id)
);
INSERT INTO meeting_vote_receipts(vote_id,voter_id)
 SELECT DISTINCT vote_id,voter_id FROM meeting_vote_records WHERE voter_id IS NOT NULL;
ALTER TABLE meeting_vote_records ALTER COLUMN voter_id DROP NOT NULL;
ALTER TABLE meeting_vote_records ALTER COLUMN voted_at DROP NOT NULL;
ALTER TABLE meeting_vote_records ALTER COLUMN voted_at DROP DEFAULT;
UPDATE meeting_vote_records r SET voter_id=NULL,voted_at=NULL
 FROM meeting_votes v WHERE v.id=r.vote_id AND v.is_anonymous;

-- Audit still records who attempted a vote and whether it succeeded, but never
-- their selected option (including old logs from the previous implementation).
UPDATE audit_logs SET request_params='[redacted: ballot choice]',
 before_json='',after_json='',diff_json=''
 WHERE path LIKE '/api/v1/votes/%/cast';
