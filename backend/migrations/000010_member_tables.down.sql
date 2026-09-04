DROP TABLE IF EXISTS member_profiles;
DROP TABLE IF EXISTS member_applications;

ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_position;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_department;
