DELETE FROM permissions WHERE code IN ('interview:manage', 'interview:evaluate');

DROP INDEX IF EXISTS uq_interview_eval_dim;
ALTER TABLE interview_evaluations
    DROP COLUMN IF EXISTS dimension,
    DROP COLUMN IF EXISTS updated_at;
CREATE UNIQUE INDEX IF NOT EXISTS interview_evaluations_interview_id_interviewer_id_key
    ON interview_evaluations (interview_id, interviewer_id);

DROP INDEX IF EXISTS idx_interviews_applicant_id;
DROP INDEX IF EXISTS idx_interviews_session_id;
ALTER TABLE interviews
    DROP COLUMN IF EXISTS session_id,
    DROP COLUMN IF EXISTS applicant_id,
    DROP COLUMN IF EXISTS actual_start_time,
    DROP COLUMN IF EXISTS actual_end_time,
    DROP COLUMN IF EXISTS result_code,
    DROP COLUMN IF EXISTS result_comment;

DROP TABLE IF EXISTS interview_dimensions;
DROP TABLE IF EXISTS interview_sessions;
