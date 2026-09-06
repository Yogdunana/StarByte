DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE code IN ('form:read', 'form:write', 'form:submit'));

DELETE FROM permissions WHERE code IN ('form:read', 'form:write', 'form:submit');

DROP TABLE IF EXISTS form_submissions;
DROP TABLE IF EXISTS forms;
