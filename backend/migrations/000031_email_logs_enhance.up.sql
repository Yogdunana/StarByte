-- #25 邮件发送记录增强：关联通知、重试次数、抄送
ALTER TABLE email_logs
    ADD COLUMN IF NOT EXISTS notification_id UUID REFERENCES notifications(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS retry_count SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cc VARCHAR(1000),
    ADD COLUMN IF NOT EXISTS is_html BOOLEAN NOT NULL DEFAULT false;

CREATE INDEX IF NOT EXISTS idx_email_logs_notification_id ON email_logs(notification_id);
