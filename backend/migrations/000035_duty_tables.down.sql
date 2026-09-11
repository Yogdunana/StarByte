-- 000035_duty_tables.down.sql
-- Issue #53: 值班/排班管理

BEGIN;

DROP TABLE IF EXISTS duty_swap_requests;
DROP TABLE IF EXISTS duty_schedules;

COMMIT;
