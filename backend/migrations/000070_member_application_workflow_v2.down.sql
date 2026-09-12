-- Keep the member_application definition (000061 / 000070). Rolling back only
-- unpublishes versions created as v2+ when they still match the v2 default graph.

UPDATE flow_definition_versions v
SET status = 0
FROM flow_definitions d
WHERE v.definition_id = d.id
  AND d.key = 'member_application'
  AND v.status = 1
  AND jsonb_array_length(v.bpmn_data->'nodes') = 5
  AND v.bpmn_data #>> '{nodes,1,id}' = 'officer';

UPDATE flow_definition_versions v
SET status = 1
FROM flow_definitions d
WHERE v.definition_id = d.id
  AND d.key = 'member_application'
  AND v.status = 0
  AND jsonb_array_length(v.bpmn_data->'nodes') = 4
  AND NOT EXISTS (
    SELECT 1 FROM flow_definition_versions x
    WHERE x.definition_id = d.id AND x.status = 1
  );
