DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE code = 'interview_private:read');
DELETE FROM permissions WHERE code = 'interview_private:read';
-- Deliberately retain the privacy-safe notification template on rollback.

-- Retain the restricted archive and sanitized public copies on rollback.
-- Restoring confidential text to candidate-visible records is never automatic.
