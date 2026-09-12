-- #65 remaining: transfer signature tables + default handover graphs.
-- Idempotent definition seed so later designer edits are not overwritten.

CREATE TABLE IF NOT EXISTS task_transfers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    from_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    to_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    initiator_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    source_department_id UUID,
    target_department_id UUID,
    source_center_id UUID,
    target_center_id UUID,
    supervisor_role VARCHAR(32) NOT NULL DEFAULT 'minister',
    kind VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    revision BIGINT NOT NULL DEFAULT 1,
    reason TEXT NOT NULL,
    previous_status SMALLINT NOT NULL DEFAULT 0,
    workflow_instance_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_task_transfers_task_status ON task_transfers(task_id, status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_task_transfers_one_pending ON task_transfers(task_id) WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS task_transfer_signatures (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transfer_id UUID NOT NULL REFERENCES task_transfers(id) ON DELETE CASCADE,
    requirement VARCHAR(32) NOT NULL,
    signer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    signer_role VARCHAR(32) NOT NULL,
    waived BOOLEAN NOT NULL DEFAULT FALSE,
    decision VARCHAR(16) NOT NULL,
    comment TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_task_transfer_signatures_transfer ON task_transfer_signatures(transfer_id, created_at);

CREATE OR REPLACE FUNCTION starbyte_seed_task_transfer_definition(p_key TEXT, p_name TEXT, p_description TEXT, p_graph JSONB)
RETURNS VOID AS $$
DECLARE
  target_definition_id uuid;
BEGIN
  INSERT INTO flow_definitions(id, key, name, description, category, status, created_at, updated_at)
  VALUES (uuid_generate_v4(), p_key, p_name, p_description, 'task', 0, NOW(), NOW())
  ON CONFLICT (key) DO NOTHING;

  SELECT id INTO target_definition_id FROM flow_definitions WHERE key = p_key FOR UPDATE;

  IF NOT EXISTS (
    SELECT 1 FROM flow_definition_versions
    WHERE definition_id = target_definition_id AND status = 1
  ) THEN
    INSERT INTO flow_definition_versions(id, definition_id, version, bpmn_data, status, published_at, created_at)
    VALUES (uuid_generate_v4(), target_definition_id, 1, p_graph, 1, NOW(), NOW());
    UPDATE flow_definitions SET status = 1, updated_at = NOW() WHERE id = target_definition_id;
  END IF;
END;
$$ LANGUAGE plpgsql;

SELECT starbyte_seed_task_transfer_definition(
  'task_transfer_internal',
  '同部门任务委托',
  '同部门委托默认直达接手人。该定义保留给需要负责人确认的内部转办。',
  $json${
    "nodes": [
      {"id": "start", "type": "start", "position": {"x": 280, "y": 0}, "data": {"label": "申请转办", "config": {}}},
      {"id": "supervisor", "type": "approval", "position": {"x": 280, "y": 160}, "data": {"label": "部门负责人确认", "config": {"assigneeStrategy": "business_role", "businessType": "task_transfer", "transferStage": "handover", "transferRole": "supervisor", "approvalType": "any", "dueDays": 1, "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
      {"id": "end", "type": "end", "position": {"x": 280, "y": 320}, "data": {"label": "转办完成", "config": {}}}
    ],
    "edges": [
      {"id": "start-supervisor", "source": "start", "target": "supervisor"},
      {"id": "supervisor-end", "source": "supervisor", "target": "end"}
    ],
    "viewport": {"x": 0, "y": 0, "zoom": 1}
  }$json$::jsonb
);

SELECT starbyte_seed_task_transfer_definition(
  'task_transfer_department',
  '跨部门任务转办',
  '跨部门转办须双方部长签字。可在流程设计器调整标签与布局，不可删减签字节点。',
  $json${
    "nodes": [
      {"id": "start", "type": "start", "position": {"x": 280, "y": 0}, "data": {"label": "申请转办", "config": {}}},
      {"id": "fork", "type": "parallel_gateway", "position": {"x": 280, "y": 140}, "data": {"label": "fork", "config": {}}},
      {"id": "source_minister", "type": "approval", "position": {"x": 80, "y": 280}, "data": {"label": "转出部门部长", "config": {"assigneeStrategy": "business_role", "businessType": "task_transfer", "transferStage": "handover", "transferRole": "source_minister", "approvalType": "any", "dueDays": 1, "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
      {"id": "target_minister", "type": "approval", "position": {"x": 480, "y": 280}, "data": {"label": "转入部门部长", "config": {"assigneeStrategy": "business_role", "businessType": "task_transfer", "transferStage": "handover", "transferRole": "target_minister", "approvalType": "any", "dueDays": 1, "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
      {"id": "join", "type": "parallel_gateway", "position": {"x": 280, "y": 420}, "data": {"label": "join", "config": {}}},
      {"id": "end", "type": "end", "position": {"x": 280, "y": 560}, "data": {"label": "转办完成", "config": {}}}
    ],
    "edges": [
      {"id": "start-fork", "source": "start", "target": "fork"},
      {"id": "fork-source_minister", "source": "fork", "target": "source_minister"},
      {"id": "fork-target_minister", "source": "fork", "target": "target_minister"},
      {"id": "source_minister-join", "source": "source_minister", "target": "join"},
      {"id": "target_minister-join", "source": "target_minister", "target": "join"},
      {"id": "join-end", "source": "join", "target": "end"}
    ],
    "viewport": {"x": 0, "y": 0, "zoom": 1}
  }$json$::jsonb
);

SELECT starbyte_seed_task_transfer_definition(
  'task_transfer_center',
  '跨中心任务转办',
  '跨中心转办须双方部长与中心负责人签字。可在流程设计器调整标签与布局，不可删减签字节点。',
  $json${
    "nodes": [
      {"id": "start", "type": "start", "position": {"x": 280, "y": 0}, "data": {"label": "申请转办", "config": {}}},
      {"id": "fork", "type": "parallel_gateway", "position": {"x": 280, "y": 140}, "data": {"label": "fork", "config": {}}},
      {"id": "source_minister", "type": "approval", "position": {"x": 80, "y": 280}, "data": {"label": "转出部门部长", "config": {"assigneeStrategy": "business_role", "businessType": "task_transfer", "transferStage": "handover", "transferRole": "source_minister", "approvalType": "any", "dueDays": 1, "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
      {"id": "target_minister", "type": "approval", "position": {"x": 280, "y": 280}, "data": {"label": "转入部门部长", "config": {"assigneeStrategy": "business_role", "businessType": "task_transfer", "transferStage": "handover", "transferRole": "target_minister", "approvalType": "any", "dueDays": 1, "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
      {"id": "source_center", "type": "approval", "position": {"x": 480, "y": 280}, "data": {"label": "转出中心负责人", "config": {"assigneeStrategy": "business_role", "businessType": "task_transfer", "transferStage": "handover", "transferRole": "source_center", "approvalType": "any", "dueDays": 1, "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
      {"id": "target_center", "type": "approval", "position": {"x": 680, "y": 280}, "data": {"label": "转入中心负责人", "config": {"assigneeStrategy": "business_role", "businessType": "task_transfer", "transferStage": "handover", "transferRole": "target_center", "approvalType": "any", "dueDays": 1, "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
      {"id": "join", "type": "parallel_gateway", "position": {"x": 280, "y": 420}, "data": {"label": "join", "config": {}}},
      {"id": "end", "type": "end", "position": {"x": 280, "y": 560}, "data": {"label": "转办完成", "config": {}}}
    ],
    "edges": [
      {"id": "start-fork", "source": "start", "target": "fork"},
      {"id": "fork-source_minister", "source": "fork", "target": "source_minister"},
      {"id": "fork-target_minister", "source": "fork", "target": "target_minister"},
      {"id": "fork-source_center", "source": "fork", "target": "source_center"},
      {"id": "fork-target_center", "source": "fork", "target": "target_center"},
      {"id": "source_minister-join", "source": "source_minister", "target": "join"},
      {"id": "target_minister-join", "source": "target_minister", "target": "join"},
      {"id": "source_center-join", "source": "source_center", "target": "join"},
      {"id": "target_center-join", "source": "target_center", "target": "join"},
      {"id": "join-end", "source": "join", "target": "end"}
    ],
    "viewport": {"x": 0, "y": 0, "zoom": 1}
  }$json$::jsonb
);

DROP FUNCTION IF EXISTS starbyte_seed_task_transfer_definition(TEXT, TEXT, TEXT, JSONB);

UPDATE flow_definitions
SET description = '默认任务链：发布 → 认领/分配 → 执行 → 审核 → 验收 → 完成。支持部门/角色/轮询自动分配、超时升级与委托转办。可在流程设计器调整标签与布局。',
    updated_at = NOW()
WHERE key = 'task_lifecycle';
