-- ============================================================
-- 000007_workflow_draft_graph.up.sql
-- 流程定义增加服务端草稿图，供设计器保存未发布版本
-- ============================================================

ALTER TABLE flow_definitions ADD COLUMN IF NOT EXISTS draft_graph JSONB;
