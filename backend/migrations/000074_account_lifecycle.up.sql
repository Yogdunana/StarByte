-- CAS/local registration writes member_profiles via GORM Create. A nil jsonb
-- slice is sent as NULL and bypasses the 000020 column default, so new
-- registrations fail with SQLSTATE 23502 on skills/projects.
-- Also: soft-deleted users kept occupying student numbers and identities;
-- hard delete was blocked by audit_logs / flow_instances FKs; new accounts
-- must verify email before login.

UPDATE member_profiles SET skills = '[]'::jsonb WHERE skills IS NULL;
UPDATE member_profiles SET projects = '[]'::jsonb WHERE projects IS NULL;
ALTER TABLE member_profiles ALTER COLUMN skills SET DEFAULT '[]'::jsonb;
ALTER TABLE member_profiles ALTER COLUMN skills SET NOT NULL;
ALTER TABLE member_profiles ALTER COLUMN projects SET DEFAULT '[]'::jsonb;
ALTER TABLE member_profiles ALTER COLUMN projects SET NOT NULL;

ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ;
UPDATE users
SET email_verified_at = COALESCE(last_login_at, created_at)
WHERE deleted_at IS NULL AND email_verified_at IS NULL;

CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_user_id
    ON email_verification_tokens(user_id);

-- Live usernames must be reusable after soft delete.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
CREATE UNIQUE INDEX IF NOT EXISTS uk_users_username_live
    ON users (username) WHERE deleted_at IS NULL;

-- Audit rows outlive the actor; keep the log, drop the FK on delete.
ALTER TABLE audit_logs ALTER COLUMN user_id DROP NOT NULL;
DO $$
DECLARE r record;
BEGIN
    FOR r IN
        SELECT c.conname
        FROM pg_constraint c
        JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY (c.conkey)
        WHERE c.contype = 'f'
          AND c.conrelid = 'audit_logs'::regclass
          AND c.confrelid = 'users'::regclass
          AND a.attname = 'user_id'
    LOOP
        EXECUTE format('ALTER TABLE audit_logs DROP CONSTRAINT %I', r.conname);
    END LOOP;
END $$;
ALTER TABLE audit_logs
    ADD CONSTRAINT audit_logs_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- Historical flow instances keep running after the initiator account is removed.
ALTER TABLE flow_instances ALTER COLUMN initiator_id DROP NOT NULL;
DO $$
DECLARE r record;
BEGIN
    FOR r IN
        SELECT c.conname
        FROM pg_constraint c
        JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY (c.conkey)
        WHERE c.contype = 'f'
          AND c.conrelid = 'flow_instances'::regclass
          AND c.confrelid = 'users'::regclass
          AND a.attname = 'initiator_id'
    LOOP
        EXECUTE format('ALTER TABLE flow_instances DROP CONSTRAINT %I', r.conname);
    END LOOP;
END $$;
ALTER TABLE flow_instances
    ADD CONSTRAINT flow_instances_initiator_id_fkey
    FOREIGN KEY (initiator_id) REFERENCES users(id) ON DELETE SET NULL;
