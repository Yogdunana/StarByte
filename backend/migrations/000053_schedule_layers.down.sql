-- ============================================================
-- 000053_schedule_layers.down.sql
-- ============================================================

DROP TABLE IF EXISTS schedule_google_accounts;

DROP INDEX IF EXISTS idx_schedule_events_origin;
ALTER TABLE schedule_events DROP COLUMN IF EXISTS external_uid;
ALTER TABLE schedule_events DROP COLUMN IF EXISTS origin;

DROP INDEX IF EXISTS idx_calendars_source;
DROP INDEX IF EXISTS idx_calendars_owner_system_layer;
DROP INDEX IF EXISTS idx_calendars_personal_owner;
ALTER TABLE calendars DROP COLUMN IF EXISTS source_key;
ALTER TABLE calendars DROP COLUMN IF EXISTS source;

CREATE UNIQUE INDEX IF NOT EXISTS idx_calendars_personal_owner
    ON calendars(owner_id) WHERE calendar_type = 1;
