DELETE FROM permissions WHERE code IN ('member:approve', 'member:export', 'member:manage');

DROP TABLE IF EXISTS member_profile_histories;
DROP TABLE IF EXISTS member_application_histories;

DROP INDEX IF EXISTS uq_member_profiles_student_no;

ALTER TABLE member_profiles
    DROP COLUMN IF EXISTS real_name,
    DROP COLUMN IF EXISTS student_no,
    DROP COLUMN IF EXISTS gender,
    DROP COLUMN IF EXISTS grade,
    DROP COLUMN IF EXISTS major,
    DROP COLUMN IF EXISTS leave_date,
    DROP COLUMN IF EXISTS skills,
    DROP COLUMN IF EXISTS projects,
    DROP COLUMN IF EXISTS bio,
    DROP COLUMN IF EXISTS contact_phone,
    DROP COLUMN IF EXISTS contact_email;

DROP INDEX IF EXISTS idx_member_applications_type;
DROP INDEX IF EXISTS idx_member_applications_student_no;

ALTER TABLE member_applications
    DROP COLUMN IF EXISTS real_name,
    DROP COLUMN IF EXISTS student_no,
    DROP COLUMN IF EXISTS skills,
    DROP COLUMN IF EXISTS experience,
    DROP COLUMN IF EXISTS contact_phone,
    DROP COLUMN IF EXISTS contact_email,
    DROP COLUMN IF EXISTS flow_instance_id,
    DROP COLUMN IF EXISTS review_comment,
    DROP COLUMN IF EXISTS required_fields;
