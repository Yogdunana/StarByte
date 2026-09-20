-- Align stored role/position titles with charter product names.
-- Internal codes (president, probationary, officer, super_admin) stay unchanged.

UPDATE roles SET name = v.name, description = v.description
FROM (VALUES
  ('super_admin', '系统管理员', '系统内置超管，不是协会会长'),
  ('president', '会长', '协会会长'),
  ('vice_president', '副会长', '协会副会长'),
  ('officer', '正式干事', '部门正式干事'),
  ('probationary', '预备成员', '预备期成员（会员通道为预备会员，干事通道为预备干事）'),
  ('member', '会员', '普通会员，不隶属部门'),
  ('vice_minister', '副部长', '部门副部长')
) AS v(code, name, description)
WHERE roles.code = v.code;

UPDATE positions SET name = v.name
FROM (VALUES
  ('president', '会长'),
  ('vice_president', '副会长'),
  ('officer', '正式干事')
) AS v(code, name)
WHERE positions.code = v.code;

UPDATE departments
SET description = replace(description, '副社长', '副会长'),
    updated_at = CURRENT_TIMESTAMP
WHERE description LIKE '%副社长%';

UPDATE flow_definition_versions
SET bpmn_data = replace(bpmn_data::text, '"label": "社长审批"', '"label": "会长审批"')::jsonb
WHERE bpmn_data::text LIKE '%"label": "社长审批"%';
