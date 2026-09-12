DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE code IN (
        'feature:read', 'feature:create', 'feature:update', 'feature:manage'
    )
);

DELETE FROM permissions
WHERE code IN ('feature:read', 'feature:create', 'feature:update', 'feature:manage');

DROP TABLE IF EXISTS feature_flag_audits;
DROP TABLE IF EXISTS feature_flags;
