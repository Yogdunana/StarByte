-- Membership approval template v2 (#64):
-- 干事审批 → 部长审批 → 社长审批，可转交；会员申请跳过干事，无部门跳过部长。
-- Only replace the published version when it is still the phase-1 default
-- (4 nodes, allowTransfer=false). Designer-published graphs are left alone.

DO $migration$
DECLARE
  target_definition_id uuid;
  next_version integer;
  graph jsonb := $json${
    "nodes": [
      {"id": "start", "type": "start", "position": {"x": 280, "y": 20}, "data": {"label": "提交申请", "config": {}}},
      {"id": "officer", "type": "approval", "position": {"x": 280, "y": 160}, "data": {"label": "干事审批", "config": {"assigneeStrategy": "role", "roleCode": "officer", "approvalType": "any", "departmentScope": true, "allowReject": true, "allowTransfer": true, "allowRollback": false, "skipWhen": "skip_officer", "skipIfEmpty": true}}},
      {"id": "minister", "type": "approval", "position": {"x": 280, "y": 300}, "data": {"label": "部长审批", "config": {"assigneeStrategy": "role", "roleCode": "minister", "approvalType": "any", "departmentScope": true, "allowReject": true, "allowTransfer": true, "allowRollback": false, "skipWhen": "skip_minister", "skipIfEmpty": true}}},
      {"id": "president", "type": "approval", "position": {"x": 280, "y": 440}, "data": {"label": "社长审批", "config": {"assigneeStrategy": "role", "roleCode": "president", "approvalType": "any", "allowReject": true, "allowTransfer": true, "allowRollback": false}}},
      {"id": "end", "type": "end", "position": {"x": 280, "y": 580}, "data": {"label": "结束", "config": {}}}
    ],
    "edges": [
      {"id": "start-officer", "source": "start", "target": "officer"},
      {"id": "officer-minister", "source": "officer", "target": "minister"},
      {"id": "minister-president", "source": "minister", "target": "president"},
      {"id": "president-end", "source": "president", "target": "end"}
    ],
    "viewport": {"x": 0, "y": 0, "zoom": 1}
  }$json$::jsonb;
BEGIN
  INSERT INTO flow_definitions(id, key, name, description, category, status, created_at, updated_at)
  VALUES (
    uuid_generate_v4(),
    'member_application',
    '入会申请审批',
    '默认可配置链：干事审批 → 部长审批 → 社长审批。支持或签/会签、条件分支与转交；可在流程设计器增删节点后发布新版本。',
    'member',
    0,
    NOW(),
    NOW()
  )
  ON CONFLICT (key) DO UPDATE SET
    description = EXCLUDED.description,
    updated_at = NOW();

  SELECT id INTO target_definition_id FROM flow_definitions WHERE key = 'member_application' FOR UPDATE;

  IF EXISTS (
    SELECT 1 FROM flow_definition_versions v
    WHERE v.definition_id = target_definition_id AND v.status = 1
      AND jsonb_array_length(v.bpmn_data->'nodes') = 4
      AND COALESCE((v.bpmn_data #>> '{nodes,1,data,config,allowTransfer}')::boolean, false) = false
  ) THEN
    SELECT COALESCE(MAX(version), 0) + 1 INTO next_version
    FROM flow_definition_versions WHERE definition_id = target_definition_id;
    UPDATE flow_definition_versions SET status = 0
    WHERE definition_id = target_definition_id AND status = 1;
    INSERT INTO flow_definition_versions(id, definition_id, version, bpmn_data, status, published_at, created_at)
    VALUES (uuid_generate_v4(), target_definition_id, next_version, graph, 1, NOW(), NOW());
    UPDATE flow_definitions SET status = 1, updated_at = NOW() WHERE id = target_definition_id;
  ELSIF NOT EXISTS (
    SELECT 1 FROM flow_definition_versions
    WHERE definition_id = target_definition_id AND status = 1
  ) THEN
    INSERT INTO flow_definition_versions(id, definition_id, version, bpmn_data, status, published_at, created_at)
    VALUES (uuid_generate_v4(), target_definition_id, 1, graph, 1, NOW(), NOW());
    UPDATE flow_definitions SET status = 1, updated_at = NOW() WHERE id = target_definition_id;
  END IF;
END $migration$;
