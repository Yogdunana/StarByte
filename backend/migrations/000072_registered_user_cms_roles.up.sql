-- Registration is separate from association membership. Existing memberships remain unchanged.
INSERT INTO roles (id, name, code, description, status, is_system)
VALUES (gen_random_uuid(), 'User', 'user', 'Registered account; association membership requires approval', 0, true)
ON CONFLICT (code) DO NOTHING;
INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT gen_random_uuid(), r.id, p.id, 'self' FROM roles r CROSS JOIN permissions p
WHERE r.code = 'user' AND p.code = 'announcement:read'
ON CONFLICT (role_id, permission_id) DO NOTHING;
ALTER TABLE knowledge_docs ADD COLUMN allowed_roles JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE knowledge_docs ADD CONSTRAINT knowledge_allowed_roles_array CHECK (jsonb_typeof(allowed_roles) = 'array');
ALTER TABLE knowledge_docs DROP CONSTRAINT chk_knowledge_docs_visibility;
ALTER TABLE knowledge_docs ADD CONSTRAINT chk_knowledge_docs_visibility CHECK (visibility IN ('public','authenticated','permission','role'));

-- Unambiguously unassigned registrations can safely receive the base role.
INSERT INTO user_roles (id, user_id, role_id)
SELECT gen_random_uuid(), u.id, r.id FROM users u CROSS JOIN roles r
WHERE r.code = 'user' AND u.deleted_at IS NULL
AND NOT EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id)
AND NOT EXISTS (SELECT 1 FROM member_profiles mp WHERE mp.user_id = u.id)
ON CONFLICT (user_id, role_id) DO NOTHING;
