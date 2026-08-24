-- ============================================================
-- 000002_workflow_engine.down.sql
-- 回滚流程引擎表
-- ============================================================

DROP TABLE IF EXISTS flow_variables;
DROP TABLE IF EXISTS flow_histories;
DROP TABLE IF EXISTS flow_tasks;
DROP TABLE IF EXISTS flow_instances;
DROP TABLE IF EXISTS flow_definition_versions;
DROP TABLE IF EXISTS flow_definitions;
