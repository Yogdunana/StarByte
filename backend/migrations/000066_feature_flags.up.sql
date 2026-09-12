-- ============================================================
-- 000066_feature_flags.up.sql
-- Issue #98 phase-1：特性开关 / 灰度（boolean / 用户名单 / 角色部门 / 百分比）
-- 错误码 34000-34999（33000 留给 #180 知识库/CMS，000064/000065 预留给 #180/#176）
-- 权限授给 president / super_admin / vice_president
-- ============================================================

CREATE TABLE IF NOT EXISTS feature_flags (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    flag_key        VARCHAR(128) NOT NULL,
    name            VARCHAR(128) NOT NULL,
    description     TEXT         NOT NULL DEFAULT '',
    flag_type       VARCHAR(32)  NOT NULL,
    enabled         BOOLEAN      NOT NULL DEFAULT FALSE,
    group_name      VARCHAR(64)  NOT NULL DEFAULT '',
    priority        INT          NOT NULL DEFAULT 0,
    rules           JSONB        NOT NULL DEFAULT '{}'::jsonb,
    is_system       BOOLEAN      NOT NULL DEFAULT FALSE,
    created_by      UUID,
    updated_by      UUID,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_feature_flags_key UNIQUE (flag_key),
    CONSTRAINT chk_feature_flags_type CHECK (
        flag_type IN ('boolean', 'user_allowlist', 'role_dept', 'percentage')
    )
);

CREATE INDEX IF NOT EXISTS idx_feature_flags_group
    ON feature_flags (group_name, priority DESC);

CREATE TABLE IF NOT EXISTS feature_flag_audits (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    flag_id     UUID,
    flag_key    VARCHAR(128) NOT NULL,
    action      VARCHAR(32)  NOT NULL,
    actor_id    UUID,
    before_json JSONB,
    after_json  JSONB,
    reason      TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feature_flag_audits_created
    ON feature_flag_audits (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_feature_flag_audits_key
    ON feature_flag_audits (flag_key, created_at DESC);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '特性开关查看', 'feature:read', 'feature', 'read', '查看特性开关、评估结果与审计', 3, true, 0),
    (uuid_generate_v4(), '特性开关创建', 'feature:create', 'feature', 'create', '创建特性开关', 3, true, 0),
    (uuid_generate_v4(), '特性开关更新', 'feature:update', 'feature', 'update', '更新特性开关规则', 3, true, 0),
    (uuid_generate_v4(), '特性开关管理', 'feature:manage', 'feature', 'manage', '切换开关与热更新发布', 3, true, 0)
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'all'
FROM roles r
CROSS JOIN permissions p
WHERE r.code IN ('president', 'super_admin', 'vice_president')
  AND p.code IN ('feature:read', 'feature:create', 'feature:update', 'feature:manage')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 一期种子：新表面默认关；已上线的公告信息流默认开，可在招新前改成名单/百分比
INSERT INTO feature_flags (id, flag_key, name, description, flag_type, enabled, group_name, priority, rules, is_system)
VALUES
    (
        '00000000-0000-4000-8000-000000000166',
        'cms.public',
        '公开 CMS',
        '公开 CMS / 已发布内容页。默认关闭，仅允许名单/角色/百分比灰度。',
        'boolean',
        FALSE,
        'cms',
        10,
        '{}'::jsonb,
        TRUE
    ),
    (
        '00000000-0000-4000-8000-000000000167',
        'announcement.feed',
        '公告信息流',
        '工作台最新公告与公告列表（成员侧）。默认开启以免打断已上线公告；招新前可改为用户名单或百分比。',
        'boolean',
        TRUE,
        'announcement',
        20,
        '{}'::jsonb,
        TRUE
    ),
    (
        '00000000-0000-4000-8000-000000000168',
        'membership.portal',
        '会员门户',
        '新的会员门户页（不替代入会申请）。默认关闭。',
        'boolean',
        FALSE,
        'membership',
        30,
        '{}'::jsonb,
        TRUE
    )
ON CONFLICT (flag_key) DO NOTHING;
