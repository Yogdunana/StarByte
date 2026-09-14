ALTER TABLE flow_instances DROP CONSTRAINT IF EXISTS flow_instances_initiator_id_fkey;
UPDATE flow_instances SET initiator_id = (
    SELECT id FROM users ORDER BY created_at LIMIT 1
) WHERE initiator_id IS NULL;
ALTER TABLE flow_instances ALTER COLUMN initiator_id SET NOT NULL;
ALTER TABLE flow_instances
    ADD CONSTRAINT flow_instances_initiator_id_fkey
    FOREIGN KEY (initiator_id) REFERENCES users(id);

ALTER TABLE audit_logs DROP CONSTRAINT IF EXISTS audit_logs_user_id_fkey;
ALTER TABLE audit_logs
    ADD CONSTRAINT audit_logs_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id);

DROP INDEX IF EXISTS uk_users_username_live;
ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);

DROP TABLE IF EXISTS email_verification_tokens;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
