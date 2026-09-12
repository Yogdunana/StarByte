-- ============================================================
-- 000056_announcement.up.sql
-- Issue #77 phase-1：公告与阅读回执（发布到全体成员）
-- 错误码使用 28000-28999（issue 写的 10000-10499 已被实习模块占用）
-- ============================================================

CREATE TABLE IF NOT EXISTS announcements (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title         VARCHAR(200) NOT NULL,
    content       TEXT NOT NULL DEFAULT '',
    content_type  VARCHAR(16) NOT NULL DEFAULT 'markdown', -- markdown | html
    category      VARCHAR(32) NOT NULL, -- association / activity / system / personnel
    pinned        BOOLEAN NOT NULL DEFAULT FALSE,
    required      BOOLEAN NOT NULL DEFAULT FALSE,
    status        SMALLINT NOT NULL DEFAULT 0, -- 0=草稿 1=已发布 2=已归档
    scheduled_at  TIMESTAMPTZ,
    published_at  TIMESTAMPTZ,
    author_id     UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at    TIMESTAMPTZ,
    CONSTRAINT chk_announcements_category CHECK (category IN ('association', 'activity', 'system', 'personnel')),
    CONSTRAINT chk_announcements_content_type CHECK (content_type IN ('markdown', 'html')),
    CONSTRAINT chk_announcements_status CHECK (status IN (0, 1, 2))
);

COMMENT ON COLUMN announcements.category IS 'association=协会公告 activity=活动通知 system=系统通知 personnel=人事变动';
COMMENT ON COLUMN announcements.status IS '0=草稿 1=已发布 2=已归档';
COMMENT ON COLUMN announcements.content_type IS 'markdown 或 html；phase-1 以正文为主，附件上传后续';
COMMENT ON COLUMN announcements.required IS '必读标记；阅读时长统计后续';

CREATE INDEX IF NOT EXISTS idx_announcements_status ON announcements(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_announcements_category ON announcements(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_announcements_pinned ON announcements(pinned, published_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_announcements_author ON announcements(author_id);
CREATE INDEX IF NOT EXISTS idx_announcements_scheduled
    ON announcements(scheduled_at)
    WHERE deleted_at IS NULL AND status = 0 AND scheduled_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS announcement_reads (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    read_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (announcement_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_announcement_reads_user ON announcement_reads(user_id);
CREATE INDEX IF NOT EXISTS idx_announcement_reads_announcement ON announcement_reads(announcement_id);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '公告查看', 'announcement:read', 'announcement', 'read', '查看公告与阅读回执', 3, true, 0),
    (uuid_generate_v4(), '公告创建', 'announcement:create', 'announcement', 'create', '创建公告草稿', 3, true, 0),
    (uuid_generate_v4(), '公告更新', 'announcement:update', 'announcement', 'update', '更新公告', 3, true, 0),
    (uuid_generate_v4(), '公告删除', 'announcement:delete', 'announcement', 'delete', '删除公告', 3, true, 0),
    (uuid_generate_v4(), '公告发布', 'announcement:publish', 'announcement', 'publish', '发布公告', 3, true, 0),
    (uuid_generate_v4(), '公告管理', 'announcement:manage', 'announcement', 'manage', '置顶与归档公告', 3, true, 0)
ON CONFLICT (code) DO NOTHING;

-- 所有登录角色可看已发布公告（data_scope=all，避免 self 把别人的公告滤掉）
INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'all'
FROM roles r
CROSS JOIN permissions p
WHERE p.code = 'announcement:read'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 社长 / 超管 / 副社长：全部写权限
INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'all'
FROM roles r
CROSS JOIN permissions p
WHERE r.code IN ('president', 'super_admin', 'vice_president')
  AND p.code IN (
      'announcement:create', 'announcement:update', 'announcement:delete',
      'announcement:publish', 'announcement:manage'
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 部长：拟稿 / 改稿 / 发布（部门范围）
INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'department'
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'minister'
  AND p.code IN ('announcement:create', 'announcement:update', 'announcement:publish')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 副部长：拟稿 / 改稿
INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'department'
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'vice_minister'
  AND p.code IN ('announcement:create', 'announcement:update')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO notification_templates
    (id, code, name, title_template, body_template, channels, category, variables_schema, status)
VALUES (
    uuid_generate_v4(),
    'announcement_published',
    '公告发布',
    '新公告：{{.title}}',
    '「{{.title}}」已发布，请查看公告中心。',
    '["in_app","websocket"]',
    'announcement',
    '{"title":"string","category":"string","id":"string"}'::jsonb,
    0
)
ON CONFLICT (code) DO NOTHING;

INSERT INTO scheduler_tasks (name, code, cron_expr, timezone, handler_key, next_run_at)
SELECT '公告定时发布扫描', 'announcement_scheduled_publish', '0 * * * * *', 'Asia/Shanghai', 'announcement_scheduled_publish', NOW()
WHERE NOT EXISTS (SELECT 1 FROM scheduler_tasks WHERE code = 'announcement_scheduled_publish' AND status <> 2);
