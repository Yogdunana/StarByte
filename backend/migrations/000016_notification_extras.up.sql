-- ============================================================
-- 000016_notification_extras.up.sql
-- 通知补充：邮件发送日志、用户通知偏好
-- notification_templates 已在 000008，此处不再建表
-- Issue #19 原 000009
-- ============================================================

CREATE TABLE email_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    to_address VARCHAR(200) NOT NULL,
    subject VARCHAR(500) NOT NULL,
    template_code VARCHAR(100),
    status SMALLINT NOT NULL DEFAULT 0, -- 0=待发送 1=成功 2=失败
    error_message TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_email_logs_user_id ON email_logs(user_id);
CREATE INDEX idx_email_logs_status ON email_logs(status);
CREATE INDEX idx_email_logs_created_at ON email_logs(created_at);

CREATE TABLE notification_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    in_app_enabled BOOLEAN NOT NULL DEFAULT true,
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    websocket_enabled BOOLEAN NOT NULL DEFAULT true,
    muted_categories VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_settings_user_id ON notification_settings(user_id);
