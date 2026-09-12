\set ON_ERROR_STOP on
-- Run after migrations up. All fixtures are temporary and rolled back.
BEGIN;
CREATE TEMP TABLE flow_definitions (LIKE public.flow_definitions INCLUDING ALL);
CREATE TEMP TABLE flow_definition_versions (LIKE public.flow_definition_versions INCLUDING ALL);

-- The untouched phase-1 seed must still upgrade exactly once.
\ir ../000061_member_application_workflow.up.sql
\ir ../000070_member_application_workflow_v2.up.sql
\ir ../000070_member_application_workflow_v2.up.sql
DO $$ BEGIN
  IF (SELECT count(*) FROM flow_definition_versions) <> 2 OR
     (SELECT count(*) FROM flow_definition_versions WHERE status = 1) <> 1 OR
     NOT EXISTS (SELECT 1 FROM flow_definition_versions WHERE status = 1
                 AND bpmn_data #>> '{nodes,1,id}' = 'officer') THEN
    RAISE EXCEPTION 'Default seed did not upgrade idempotently';
  END IF;
END $$;

-- Changing the role and all-sign policy must preserve the published graph.
TRUNCATE flow_definition_versions, flow_definitions;
\ir ../000061_member_application_workflow.up.sql
UPDATE flow_definition_versions SET bpmn_data = jsonb_set(
  jsonb_set(bpmn_data, '{nodes,1,data,config,roleCode}', '"custom_reviewer"'),
  '{nodes,1,data,config,approvalType}', '"all"');
CREATE TEMP TABLE expected_graph AS SELECT * FROM flow_definition_versions;
\ir ../000070_member_application_workflow_v2.up.sql
DO $$ BEGIN
  IF (SELECT count(*) FROM flow_definition_versions) <> 1 OR
     NOT EXISTS (SELECT 1 FROM flow_definition_versions v JOIN expected_graph e
                 ON v.id = e.id WHERE v.status = 1 AND v.bpmn_data = e.bpmn_data) THEN
    RAISE EXCEPTION 'Customized approval policy was replaced';
  END IF;
END $$;

-- An edge-only change must also survive, with the same four nodes.
TRUNCATE flow_definition_versions, flow_definitions, expected_graph;
\ir ../000061_member_application_workflow.up.sql
UPDATE flow_definition_versions SET bpmn_data = jsonb_set(
  bpmn_data, '{edges,0,target}', '"president"');
INSERT INTO expected_graph SELECT * FROM flow_definition_versions;
\ir ../000070_member_application_workflow_v2.up.sql
DO $$ BEGIN
  IF (SELECT count(*) FROM flow_definition_versions) <> 1 OR
     NOT EXISTS (SELECT 1 FROM flow_definition_versions v JOIN expected_graph e
                 ON v.id = e.id WHERE v.status = 1 AND v.bpmn_data = e.bpmn_data) THEN
    RAISE EXCEPTION 'Customized routing was replaced';
  END IF;
END $$;
ROLLBACK;
