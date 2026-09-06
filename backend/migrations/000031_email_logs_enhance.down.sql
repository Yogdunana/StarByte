DROP INDEX IF EXISTS idx_email_logs_notification_id;
ALTER TABLE email_logs
    DROP COLUMN IF EXISTS notification_id,
    DROP COLUMN IF EXISTS retry_count,
    DROP COLUMN IF EXISTS cc,
    DROP COLUMN IF EXISTS is_html;
