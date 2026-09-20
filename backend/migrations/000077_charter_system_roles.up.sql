-- System offices from the charter: deputy center director, advisor, honorary member,
-- tech-team captain/teammate. Probationary is officer-track only (预备干事).
INSERT INTO roles (id, name, code, description, status, is_system, sort_order) VALUES
  (gen_random_uuid(), '副中心主任', 'vice_center_director', '中心副主任，中心范围内的协会职务', 0, true, 2),
  (gen_random_uuid(), '指导老师', 'advisor', '协会指导老师', 0, true, 9),
  (gen_random_uuid(), '荣誉会员', 'honorary', '荣誉会员，不占日常编制', 0, true, 10),
  (gen_random_uuid(), '队长', 'captain', '技术团队队长，兼职编制', 0, true, 11),
  (gen_random_uuid(), '队员', 'teammate', '技术团队队员，兼职编制', 0, true, 12)
ON CONFLICT (code) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  is_system = EXCLUDED.is_system,
  sort_order = EXCLUDED.sort_order,
  status = 0;

UPDATE roles
SET name = '预备干事',
    description = '干事通道预备期，满一个月后转为正式干事'
WHERE code = 'probationary';

INSERT INTO positions (id, name, code, level, vote_weight, sort_order, status)
VALUES (gen_random_uuid(), '副中心主任', 'vice_center_director', 8, 1, 2, 0)
ON CONFLICT (code) DO UPDATE SET
  name = EXCLUDED.name,
  vote_weight = EXCLUDED.vote_weight,
  sort_order = EXCLUDED.sort_order;

UPDATE configs
SET config_value = jsonb_set(COALESCE(config_value, '{}'::jsonb), '{weights,vice_center_director}', '1'::jsonb, true),
    updated_at = NOW()
WHERE config_key = 'vote_weight_config';

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT gen_random_uuid(), r.id, p.id, 'department_and_sub'
FROM roles r CROSS JOIN permissions p
WHERE r.code = 'vice_center_director'
  AND (p.resource IN ('member','interview','interview_private','task','workflow','meeting')
       OR p.code IN ('user:read','department:read','position:read','announcement:read','doc:read','file:read'))
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT gen_random_uuid(), r.id, p.id, 'self'
FROM roles r CROSS JOIN permissions p
WHERE r.code IN ('honorary', 'advisor', 'teammate')
  AND p.code IN (
    'user:read','member:read','meeting:read','task:read',
    'activity:read','announcement:read','doc:read',
    'file:read','file:create','internship:read','internship:create','internship:update','internship:delete',
    'schedule:read','schedule:create','schedule:update','schedule:delete',
    'form:submit','discipline:read'
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT gen_random_uuid(), r.id, p.id, 'self'
FROM roles r CROSS JOIN permissions p
WHERE r.code = 'captain'
  AND p.code IN (
    'user:read','member:read','member:create',
    'interview:read','interview:evaluate','meeting:read',
    'activity:read','activity:create','activity:update',
    'announcement:read','leave:read','doc:read',
    'task:read','task:create','task:update','task:comment',
    'file:read','file:create',
    'internship:read','internship:create','internship:update','internship:delete',
    'schedule:read','schedule:create','schedule:update','schedule:delete',
    'form:read','form:submit','discipline:read'
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT gen_random_uuid(), r.id, p.id, 'self'
FROM roles r CROSS JOIN permissions p
WHERE r.code = 'probationary'
  AND p.code IN (
    'user:read','member:read','meeting:read','task:read',
    'activity:read','announcement:read','doc:read',
    'file:read','file:create','internship:read','internship:create','internship:update','internship:delete',
    'schedule:read','schedule:create','schedule:update','schedule:delete',
    'form:submit','discipline:read'
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;
