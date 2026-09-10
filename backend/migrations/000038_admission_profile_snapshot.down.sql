-- Snapshots are evidence and must survive rollback.
DO $$ BEGIN
 IF EXISTS (SELECT 1 FROM admission_profile_snapshots) THEN
  RAISE EXCEPTION 'Admission profile snapshots exist; preserve them and use a reviewed forward migration';
 END IF;
END $$;
DROP TABLE admission_profile_snapshots;
