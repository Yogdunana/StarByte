DO $$ BEGIN
 IF EXISTS(SELECT 1 FROM meeting_votes WHERE electorate_frozen) THEN
  RAISE EXCEPTION 'Vote snapshots exist; rollback would destroy voting provenance';
 END IF;
END $$;
DROP TABLE meeting_vote_electorates;
ALTER TABLE meeting_votes DROP COLUMN weight_snapshot, DROP COLUMN eligible_count, DROP COLUMN electorate_frozen;
