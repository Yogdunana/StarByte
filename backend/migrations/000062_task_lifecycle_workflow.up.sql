-- Default task lifecycle template (#65 phase-1):
-- 发布 → 认领/分配 → 执行 → 审核 → 验收 → 完成
-- Assignees are resolved at runtime via business_role (collaboration_task).
-- Idempotent: publish only when no current version exists, so later designer
-- edits and the 000046 seed are not overwritten.

DO $migration$
DECLARE
  target_definition_id uuid;
  graph jsonb := $json${
    "nodes": [
      {"id": "start", "type": "start", "position": {"x": 280, "y": 0}, "data": {"label": "发布任务", "config": {}}},
      {"id": "assignment", "type": "approval", "position": {"x": 280, "y": 140}, "data": {"label": "认领 / 分配", "config": {"assigneeStrategy": "business_role", "businessType": "collaboration_task", "taskStage": "assignment", "approvalType": "single", "dueDays": 1, "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
      {"id": "execution", "type": "approval", "position": {"x": 280, "y": 280}, "data": {"label": "执行与提交交付", "config": {"assigneeStrategy": "business_role", "businessType": "collaboration_task", "taskStage": "execution", "approvalType": "single", "dueDays": 1, "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
      {"id": "review", "type": "approval", "position": {"x": 280, "y": 420}, "data": {"label": "负责人审核", "config": {"assigneeStrategy": "business_role", "businessType": "collaboration_task", "taskStage": "review", "approvalType": "single", "dueDays": 1, "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
      {"id": "acceptance", "type": "approval", "position": {"x": 280, "y": 560}, "data": {"label": "正式验收", "config": {"assigneeStrategy": "business_role", "businessType": "collaboration_task", "taskStage": "acceptance", "approvalType": "single", "dueDays": 1, "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
      {"id": "end", "type": "end", "position": {"x": 280, "y": 700}, "data": {"label": "任务完成", "config": {}}}
    ],
    "edges": [
      {"id": "start-assignment", "source": "start", "target": "assignment"},
      {"id": "assignment-execution", "source": "assignment", "target": "execution"},
      {"id": "execution-review", "source": "execution", "target": "review"},
      {"id": "review-acceptance", "source": "review", "target": "acceptance"},
      {"id": "acceptance-end", "source": "acceptance", "target": "end"}
    ],
    "viewport": {"x": 0, "y": 0, "zoom": 1}
  }$json$::jsonb;
BEGIN
  INSERT INTO flow_definitions(id, key, name, description, category, status, created_at, updated_at)
  VALUES (
    uuid_generate_v4(),
    'task_lifecycle',
    '任务交付与验收',
    '默认任务链：发布 → 认领/分配 → 执行 → 审核 → 验收 → 完成。可在流程设计器调整标签与布局；增删分支属后续迭代。',
    'task',
    0,
    NOW(),
    NOW()
  )
  ON CONFLICT (key) DO NOTHING;

  SELECT id INTO target_definition_id FROM flow_definitions WHERE key = 'task_lifecycle' FOR UPDATE;

  IF NOT EXISTS (
    SELECT 1 FROM flow_definition_versions
    WHERE definition_id = target_definition_id AND status = 1
  ) THEN
    INSERT INTO flow_definition_versions(id, definition_id, version, bpmn_data, status, published_at, created_at)
    VALUES (uuid_generate_v4(), target_definition_id, 1, graph, 1, NOW(), NOW());
    UPDATE flow_definitions SET status = 1, updated_at = NOW() WHERE id = target_definition_id;
  END IF;
END $migration$;
