-- ============================================================
-- 000068_leave_balance_deducted.up.sql
-- 记录提交时是否预扣余额，驳回按快照返还，避免中途改 Deductible 错账
-- ============================================================

ALTER TABLE leave_applications
    ADD COLUMN IF NOT EXISTS balance_deducted BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN leave_applications.balance_deducted IS '提交时是否已预扣年假/调休；驳回按此快照返还，不读当前类型 deductible';

UPDATE leave_applications a
SET balance_deducted = TRUE
FROM leave_types t
WHERE a.leave_type_id = t.id
  AND t.deductible = TRUE
  AND a.status = 'pending'
  AND a.deleted_at IS NULL
  AND a.balance_deducted = FALSE;

-- 已发布的默认图若仍按 role 指派，改为 business_role，避免申请人被指成唯一审批人。
UPDATE flow_definition_versions v
SET bpmn_data = $json${
  "nodes": [
    {"id": "start", "type": "start", "position": {"x": 280, "y": 20}, "data": {"label": "提交请假", "config": {}}},
    {"id": "minister", "type": "approval", "position": {"x": 280, "y": 160}, "data": {"label": "部长审批", "config": {"assigneeStrategy": "business_role", "businessType": "leave_application", "leaveStage": "department", "approvalType": "any", "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
    {"id": "president", "type": "approval", "position": {"x": 280, "y": 300}, "data": {"label": "社长审批", "config": {"assigneeStrategy": "business_role", "businessType": "leave_application", "leaveStage": "org", "approvalType": "any", "allowReject": true, "allowTransfer": false, "allowRollback": false}}},
    {"id": "end", "type": "end", "position": {"x": 280, "y": 440}, "data": {"label": "结束", "config": {}}}
  ],
  "edges": [
    {"id": "start-minister", "source": "start", "target": "minister"},
    {"id": "minister-president", "source": "minister", "target": "president"},
    {"id": "president-end", "source": "president", "target": "end"}
  ],
  "viewport": {"x": 0, "y": 0, "zoom": 1}
}$json$::jsonb
FROM flow_definitions d
WHERE v.definition_id = d.id
  AND d.key = 'leave_approval'
  AND v.status = 1
  AND v.bpmn_data::text LIKE '%"assigneeStrategy": "role"%';
