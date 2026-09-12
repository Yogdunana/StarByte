-- ============================================================
-- 000063_announcement_phase2.up.sql
-- Issue #77 剩余验收：定向受众、自动下架、附件、阅读时长、邮件渠道
-- ============================================================

ALTER TABLE announcements
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS audience_type VARCHAR(16) NOT NULL DEFAULT 'all',
    ADD COLUMN IF NOT EXISTS audience_ids JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS attachments JSONB NOT NULL DEFAULT '[]'::jsonb;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_announcements_audience_type'
    ) THEN
        ALTER TABLE announcements
            ADD CONSTRAINT chk_announcements_audience_type
            CHECK (audience_type IN ('all', 'role', 'department', 'users'));
    END IF;
END $$;

COMMENT ON COLUMN announcements.expires_at IS '到期自动归档（下架）时间';
COMMENT ON COLUMN announcements.sort_order IS '置顶后的次级排序，越小越靠前';
COMMENT ON COLUMN announcements.audience_type IS 'all=全员 role=按角色 department=按部门 users=指定用户';
COMMENT ON COLUMN announcements.audience_ids IS '受众 UUID 列表（角色/部门/用户，取决于 audience_type）';
COMMENT ON COLUMN announcements.attachments IS '附件快照 [{file_id,name,size}]';

CREATE INDEX IF NOT EXISTS idx_announcements_expires
    ON announcements(expires_at)
    WHERE deleted_at IS NULL AND status = 1 AND expires_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_announcements_sort
    ON announcements(pinned DESC, sort_order ASC, published_at DESC)
    WHERE deleted_at IS NULL;

ALTER TABLE announcement_reads
    ADD COLUMN IF NOT EXISTS duration_seconds INTEGER NOT NULL DEFAULT 0;

COMMENT ON COLUMN announcement_reads.duration_seconds IS '打开详情累计阅读秒数（首次已读后可上调）';

UPDATE notification_templates
SET channels = '["in_app","websocket","email"]'
WHERE code = 'announcement_published' AND deleted_at IS NULL;

INSERT INTO scheduler_tasks (name, code, cron_expr, timezone, handler_key, next_run_at)
SELECT '公告到期自动下架', 'announcement_auto_unpublish', '0 * * * * *', 'Asia/Shanghai', 'announcement_auto_unpublish', NOW()
WHERE NOT EXISTS (SELECT 1 FROM scheduler_tasks WHERE code = 'announcement_auto_unpublish' AND status <> 2);
