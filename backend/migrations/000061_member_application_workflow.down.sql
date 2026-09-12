-- Drop the default membership approval template only when no instances remain.

DELETE FROM flow_definition_versions
WHERE definition_id IN (
  SELECT d.id FROM flow_definitions d
  WHERE d.key = 'member_application'
    AND NOT EXISTS (SELECT 1 FROM flow_instances i WHERE i.definition_id = d.id)
);

DELETE FROM flow_definitions d
WHERE d.key = 'member_application'
  AND NOT EXISTS (SELECT 1 FROM flow_instances i WHERE i.definition_id = d.id);
