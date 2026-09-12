DELETE FROM knowledge_attachments;
DELETE FROM knowledge_doc_versions;
DELETE FROM knowledge_docs;
DELETE FROM knowledge_categories;

DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE code IN (
        'doc:read', 'doc:create', 'doc:update', 'doc:delete', 'doc:publish'
    )
);

DELETE FROM permissions
WHERE code IN (
    'doc:read', 'doc:create', 'doc:update', 'doc:delete', 'doc:publish'
);

DROP TABLE IF EXISTS knowledge_attachments;
DROP TABLE IF EXISTS knowledge_doc_versions;
DROP TABLE IF EXISTS knowledge_docs;
DROP TABLE IF EXISTS knowledge_categories;
