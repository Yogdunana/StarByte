UPDATE roles SET name = v.name, description = v.description
FROM (VALUES
  ('super_admin', '超级管理员', '系统内置超管'),
  ('president', '社长', '协会社长'),
  ('vice_president', '副社长', '协会副社长'),
  ('officer', '干事', '部门干事'),
  ('probationary', '候补成员', '入会通过后一个自然月候补期'),
  ('member', '会员', '普通会员'),
  ('vice_minister', '副部长', '部门副部长（章程未单列，系统预留）')
) AS v(code, name, description)
WHERE roles.code = v.code;

UPDATE positions SET name = v.name
FROM (VALUES
  ('president', '社长'),
  ('vice_president', '副社长'),
  ('officer', '干事')
) AS v(code, name)
WHERE positions.code = v.code;

UPDATE departments
SET description = replace(description, '副会长兼任主任', '副社长兼任主任'),
    updated_at = CURRENT_TIMESTAMP
WHERE description LIKE '%副会长兼任主任%';

UPDATE flow_definition_versions
SET bpmn_data = replace(bpmn_data::text, '"label": "会长审批"', '"label": "社长审批"')::jsonb
WHERE bpmn_data::text LIKE '%"label": "会长审批"%';
