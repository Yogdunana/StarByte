-- Identity cannot be reconstructed from anonymous ballots.
DO $$ BEGIN RAISE EXCEPTION 'Anonymous ballot separation cannot be reversed without inventing voter identities'; END $$;
