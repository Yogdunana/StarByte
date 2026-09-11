ALTER TABLE meeting_votes ADD COLUMN electorate_frozen BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE meeting_votes ADD COLUMN eligible_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE meeting_votes ADD COLUMN weight_snapshot JSONB NOT NULL DEFAULT '{}';
CREATE TABLE meeting_vote_electorates (
 vote_id UUID NOT NULL REFERENCES meeting_votes(id) ON DELETE CASCADE,
 user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
 weight NUMERIC(8,2) NOT NULL CHECK(weight>0),
 PRIMARY KEY(vote_id,user_id)
);
-- Existing polls retain electorate_frozen=false; do not infer historical roles.
