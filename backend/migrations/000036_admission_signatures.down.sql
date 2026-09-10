-- A rollback must never discard signed decisions or objections.
DO $$ BEGIN
 IF EXISTS (SELECT 1 FROM admission_signatures) OR EXISTS (SELECT 1 FROM admission_objections)
 THEN RAISE EXCEPTION 'Admission decisions exist; export and review them before rollback'; END IF;
END $$;
DROP TABLE admission_objections;
DROP TABLE admission_signatures;
ALTER TABLE member_applications DROP COLUMN admission_revision, DROP COLUMN admission_version, DROP COLUMN admission_stage,
 DROP COLUMN historical_review_required, DROP COLUMN probation_until, DROP COLUMN stage_entered_at;
