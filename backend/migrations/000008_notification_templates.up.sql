-- ============================================================
-- 000008_notification_templates.up.sql
-- 通知模板表（对齐 NotificationTemplate 模型）
-- ============================================================

CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(100) NOT NULL,
    name VARCHAR(200) NOT NULL,
    title_template VARCHAR(500) NOT NULL,
    body_template TEXT NOT NULL,
    channels VARCHAR(200) NOT NULL,
    category VARCHAR(50),
    variables_schema JSONB,
    status SMALLINT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_notification_templates_code ON notification_templates(code);
CREATE INDEX IF NOT EXISTS idx_notification_templates_status ON notification_templates(status);
CREATE INDEX IF NOT EXISTS idx_notification_templates_deleted_at ON notification_templates(deleted_at);
