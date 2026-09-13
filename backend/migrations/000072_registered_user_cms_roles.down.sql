UPDATE knowledge_docs SET visibility = 'permission', permission_code = 'doc:publish' WHERE visibility = 'role';
ALTER TABLE knowledge_docs DROP CONSTRAINT chk_knowledge_docs_visibility;
ALTER TABLE knowledge_docs ADD CONSTRAINT chk_knowledge_docs_visibility CHECK (visibility IN ('public','authenticated','permission'));
ALTER TABLE knowledge_docs DROP COLUMN IF EXISTS allowed_roles;
-- Preserve User roles and assignments: rollback must not change account membership.
