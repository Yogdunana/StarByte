-- 仅撤销本迁移写入的 schedule:* 角色授权；权限本身由 000051 持有，不删除。

DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE code IN ('schedule:read', 'schedule:create', 'schedule:update', 'schedule:delete')
);
