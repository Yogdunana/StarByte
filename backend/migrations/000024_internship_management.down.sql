DROP INDEX IF EXISTS idx_internships_start_date;
DROP INDEX IF EXISTS idx_internships_mentor_id;
DROP INDEX IF EXISTS idx_internships_type;

ALTER TABLE internships
    DROP COLUMN IF EXISTS updated_by,
    DROP COLUMN IF EXISTS created_by,
    DROP COLUMN IF EXISTS mentor_comment,
    DROP COLUMN IF EXISTS report,
    DROP COLUMN IF EXISTS achievements,
    DROP COLUMN IF EXISTS skills,
    DROP COLUMN IF EXISTS supervisor_id,
    DROP COLUMN IF EXISTS mentor_id,
    DROP COLUMN IF EXISTS type,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS organization,
    DROP COLUMN IF EXISTS title;

DELETE FROM configs WHERE config_key = 'internship_config';
DELETE FROM permissions WHERE code = 'internship:evaluate';
