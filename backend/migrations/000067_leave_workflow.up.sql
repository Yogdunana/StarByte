-- ============================================================
-- 000067_leave_workflow.up.sql
-- Issue #56：请假审批接入流程引擎、附件、类型可配置
-- 000064 knowledge / 000065 backup / 000066 feature 预留，本片使用 000067
-- ============================================================

ALTER TABLE leave_types
    ADD COLUMN IF NOT EXISTS enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN leave_types.enabled IS '关闭后不可再提交该类型，已有申请与余额保留';
COMMENT ON COLUMN leave_types.sort_order IS '申请表与余额卡片的展示顺序';

UPDATE leave_types SET sort_order = 10, enabled = TRUE WHERE code = 'annual' AND sort_order = 0;
UPDATE leave_types SET sort_order = 20, enabled = TRUE WHERE code = 'compensatory' AND sort_order = 0;
UPDATE leave_types SET sort_order = 30, enabled = TRUE WHERE code = 'personal' AND sort_order = 0;
UPDATE leave_types SET sort_order = 40, enabled = TRUE WHERE code = 'sick' AND sort_order = 0;

ALTER TABLE leave_applications
    ADD COLUMN IF NOT EXISTS workflow_instance_id UUID REFERENCES flow_instances(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS workflow_stage VARCHAR(32) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS attachments JSONB NOT NULL DEFAULT '[]'::jsonb;

CREATE INDEX IF NOT EXISTS idx_leave_applications_workflow
    ON leave_applications(workflow_instance_id)
    WHERE workflow_instance_id IS NOT NULL AND deleted_at IS NULL;

COMMENT ON COLUMN leave_applications.workflow_instance_id IS 'leave_approval 流程实例；历史 phase-1 记录可为空并走单级回退';
COMMENT ON COLUMN leave_applications.workflow_stage IS '当前引擎节点：minister / president / 空';
COMMENT ON COLUMN leave_applications.attachments IS '已上传文件快照 [{file_id,name,size}]';

-- Default leave approval template: 提交 → 部长 → 社长
-- Idempotent: publish only when no current version exists.
DO $migration$
DECLARE
  target_definition_id uuid;
  graph jsonb := $json${
    "nodes": [
      {"id": "start", "type": "start", "position": {"x": 280, "y": 20}, "data": {"label": "提交请假", "config": {}}},
      {"id": "minister", "type": "approval", "position": {"x": 280, "y": 160}, "data": {"label": "部长审批", "config": {"assigneeStrategy": "role", "roleCode": "minister", "approvalType": "any", "departmentScope": true, "leaveStage": "department", "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
      {"id": "president", "type": "approval", "position": {"x": 280, "y": 300}, "data": {"label": "社长审批", "config": {"assigneeStrategy": "role", "roleCode": "president", "approvalType": "any", "leaveStage": "org", "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
      {"id": "end", "type": "end", "position": {"x": 280, "y": 440}, "data": {"label": "结束", "config": {}}}
    ],
    "edges": [
      {"id": "start-minister", "source": "start", "target": "minister"},
      {"id": "minister-president", "source": "minister", "target": "president"},
      {"id": "president-end", "source": "president", "target": "end"}
    ],
    "viewport": {"x": 0, "y": 0, "zoom": 1}
  }$json$::jsonb;
BEGIN
  INSERT INTO flow_definitions(id, key, name, description, category, status, created_at, updated_at)
  VALUES (
    uuid_generate_v4(),
    'leave_approval',
    '请假审批',
    '默认请假链：部长审批 → 社长审批。可在流程设计器调整标签与布局；增删分支属后续迭代。',
    'leave',
    0,
    NOW(),
    NOW()
  )
  ON CONFLICT (key) DO NOTHING;

  SELECT id INTO target_definition_id FROM flow_definitions WHERE key = 'leave_approval' FOR UPDATE;

  IF NOT EXISTS (
    SELECT 1 FROM flow_definition_versions
    WHERE definition_id = target_definition_id AND status = 1
  ) THEN
    INSERT INTO flow_definition_versions(id, definition_id, version, bpmn_data, status, published_at, created_at)
    VALUES (uuid_generate_v4(), target_definition_id, 1, graph, 1, NOW(), NOW());
    UPDATE flow_definitions SET status = 1, updated_at = NOW() WHERE id = target_definition_id;
  END IF;
END $migration$;
