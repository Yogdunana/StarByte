-- Preserve prior membership when an existing member tries an officer position.
CREATE TABLE admission_profile_snapshots (
 application_id UUID PRIMARY KEY REFERENCES member_applications(id),
 profile JSONB NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
