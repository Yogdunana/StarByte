DELETE FROM role_permissions rp
USING roles r
WHERE rp.role_id = r.id
  AND r.code IN ('vice_center_director','advisor','honorary','captain','teammate');

DELETE FROM user_role_departments urd
USING user_roles ur, roles r
WHERE urd.user_role_id = ur.id AND ur.role_id = r.id
  AND r.code IN ('vice_center_director','advisor','honorary','captain','teammate');

DELETE FROM user_roles ur
USING roles r
WHERE ur.role_id = r.id
  AND r.code IN ('vice_center_director','advisor','honorary','captain','teammate');

DELETE FROM roles
WHERE code IN ('vice_center_director','advisor','honorary','captain','teammate');

DELETE FROM positions WHERE code = 'vice_center_director';

UPDATE roles
SET name = '预备成员',
    description = '预备期成员（会员通道为预备会员，干事通道为预备干事）'
WHERE code = 'probationary';

UPDATE configs
SET config_value = (config_value::jsonb #- '{weights,vice_center_director}')::text,
    updated_at = NOW()
WHERE config_key = 'vote_weight_config'
  AND NULLIF(config_value, '') IS NOT NULL
  AND config_value::jsonb #> '{weights,vice_center_director}' IS NOT NULL;
