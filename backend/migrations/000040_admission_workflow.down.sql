-- Published history is immutable. Reverting application code requires an explicit
-- reviewed forward migration; never delete real application approvals.
DO $$ BEGIN RAISE EXCEPTION 'Admission workflow versions must be preserved; use a reviewed forward migration'; END $$;
