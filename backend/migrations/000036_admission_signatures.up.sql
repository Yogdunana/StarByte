-- Existing decisions remain unchanged and are explicitly marked for review.
ALTER TABLE member_applications
 ADD COLUMN admission_revision INT NOT NULL DEFAULT 1,
 ADD COLUMN admission_version SMALLINT NOT NULL DEFAULT 1,
 ADD COLUMN admission_stage VARCHAR(40) NOT NULL DEFAULT 'legacy_review',
 ADD COLUMN historical_review_required BOOLEAN NOT NULL DEFAULT true,
 ADD COLUMN probation_until TIMESTAMPTZ,
 ADD COLUMN stage_entered_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE member_applications ALTER COLUMN admission_version SET DEFAULT 2;
ALTER TABLE member_applications ALTER COLUMN admission_stage SET DEFAULT 'materials';
ALTER TABLE member_applications ALTER COLUMN historical_review_required SET DEFAULT false;

CREATE TABLE admission_signatures (
 id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
 application_id UUID NOT NULL REFERENCES member_applications(id),
 stage VARCHAR(40) NOT NULL,
 revision INT NOT NULL DEFAULT 1,
 signer_id UUID NOT NULL REFERENCES users(id),
 signer_role VARCHAR(40) NOT NULL,
 decision VARCHAR(20) NOT NULL CHECK (decision IN ('approve','reject','supplement')),
 comment TEXT NOT NULL DEFAULT '',
 delegated BOOLEAN NOT NULL DEFAULT false,
 delegation_reason TEXT NOT NULL DEFAULT '',
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 UNIQUE(application_id, revision, stage, signer_role)
);
CREATE INDEX idx_admission_signatures_application ON admission_signatures(application_id);
CREATE INDEX idx_admission_probation_due ON member_applications(probation_until) WHERE admission_stage = 'probation';
CREATE TABLE admission_objections (
 id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
 application_id UUID NOT NULL REFERENCES member_applications(id),
 raised_by UUID NOT NULL REFERENCES users(id),
 reason TEXT NOT NULL,
 status VARCHAR(30) NOT NULL DEFAULT 'center_review',
 center_reviewer_id UUID REFERENCES users(id),
 center_comment TEXT NOT NULL DEFAULT '',
 final_reviewer_id UUID REFERENCES users(id),
 final_comment TEXT NOT NULL DEFAULT '',
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_admission_open_objection ON admission_objections(application_id) WHERE status IN ('center_review','president_review');
