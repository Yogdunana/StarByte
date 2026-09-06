-- 000028_search_engine.down.sql
DROP INDEX IF EXISTS idx_member_applications_search_fts;
DROP INDEX IF EXISTS idx_audit_logs_search_fts;
DROP INDEX IF EXISTS idx_users_real_name_trgm;
DROP INDEX IF EXISTS idx_users_search_fts;
DROP INDEX IF EXISTS idx_tasks_title_trgm;
DROP INDEX IF EXISTS idx_tasks_search_fts;
DROP FUNCTION IF EXISTS starbyte_cjk_tokens(text);
