DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE code IN ('leave:read', 'leave:approve'));

DELETE FROM permissions WHERE code IN ('leave:read', 'leave:approve');

DROP TABLE IF EXISTS leave_applications;
DROP TABLE IF EXISTS leave_balances;
DROP TABLE IF EXISTS leave_types;
