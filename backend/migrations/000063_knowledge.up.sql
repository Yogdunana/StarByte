-- ============================================================
-- 000063_knowledge.up.sql
-- Issue #58：文档 / 知识库 CMS（招新公开 about/docs + 成员手册）
-- 错误码 33000-33999
-- ============================================================

CREATE TABLE IF NOT EXISTS knowledge_categories (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_id   UUID REFERENCES knowledge_categories(id) ON DELETE SET NULL,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(80) NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_knowledge_categories_slug
    ON knowledge_categories(slug) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_knowledge_categories_parent
    ON knowledge_categories(parent_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS knowledge_docs (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    kind             VARCHAR(16) NOT NULL, -- page | doc
    slug             VARCHAR(80) NOT NULL,
    title            VARCHAR(200) NOT NULL,
    summary          VARCHAR(500) NOT NULL DEFAULT '',
    content          TEXT NOT NULL DEFAULT '',
    category_id      UUID REFERENCES knowledge_categories(id) ON DELETE SET NULL,
    visibility       VARCHAR(32) NOT NULL DEFAULT 'authenticated', -- public | authenticated | permission
    permission_code  VARCHAR(64),
    status           SMALLINT NOT NULL DEFAULT 0, -- 0=草稿 1=已发布
    version          INT NOT NULL DEFAULT 1,
    author_id        UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    published_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at       TIMESTAMPTZ,
    search_tsv       tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('simple', coalesce(summary, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(content, '')), 'C')
    ) STORED,
    CONSTRAINT chk_knowledge_docs_kind CHECK (kind IN ('page', 'doc')),
    CONSTRAINT chk_knowledge_docs_visibility CHECK (visibility IN ('public', 'authenticated', 'permission')),
    CONSTRAINT chk_knowledge_docs_status CHECK (status IN (0, 1))
);

COMMENT ON COLUMN knowledge_docs.kind IS 'page=独立页（如 /about-us） doc=/docs/:slug';
COMMENT ON COLUMN knowledge_docs.visibility IS 'public=公开 authenticated=登录可读 permission=需 permission_code';
COMMENT ON COLUMN knowledge_docs.status IS '0=草稿 1=已发布';

CREATE UNIQUE INDEX IF NOT EXISTS uq_knowledge_docs_kind_slug
    ON knowledge_docs(kind, slug) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_knowledge_docs_category
    ON knowledge_docs(category_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_knowledge_docs_status
    ON knowledge_docs(status, visibility) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_knowledge_docs_search
    ON knowledge_docs USING GIN (search_tsv);

CREATE TABLE IF NOT EXISTS knowledge_doc_versions (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    doc_id      UUID NOT NULL REFERENCES knowledge_docs(id) ON DELETE CASCADE,
    version     INT NOT NULL,
    title       VARCHAR(200) NOT NULL,
    summary     VARCHAR(500) NOT NULL DEFAULT '',
    content     TEXT NOT NULL DEFAULT '',
    editor_id   UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (doc_id, version)
);

CREATE INDEX IF NOT EXISTS idx_knowledge_doc_versions_doc
    ON knowledge_doc_versions(doc_id, version DESC);

CREATE TABLE IF NOT EXISTS knowledge_attachments (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    doc_id      UUID NOT NULL REFERENCES knowledge_docs(id) ON DELETE CASCADE,
    file_id     UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (doc_id, file_id)
);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '文档查看', 'doc:read', 'doc', 'read', '查看知识库文档', 3, true, 0),
    (uuid_generate_v4(), '文档创建', 'doc:create', 'doc', 'create', '创建知识库文档', 3, true, 0),
    (uuid_generate_v4(), '文档更新', 'doc:update', 'doc', 'update', '更新知识库文档', 3, true, 0),
    (uuid_generate_v4(), '文档删除', 'doc:delete', 'doc', 'delete', '删除知识库文档', 3, true, 0),
    (uuid_generate_v4(), '文档发布', 'doc:publish', 'doc', 'publish', '发布知识库文档', 3, true, 0)
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'all'
FROM roles r
CROSS JOIN permissions p
WHERE p.code = 'doc:read'
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'all'
FROM roles r
CROSS JOIN permissions p
WHERE r.code IN ('president', 'super_admin', 'vice_president')
  AND p.code IN ('doc:create', 'doc:update', 'doc:delete', 'doc:publish')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'department'
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'minister'
  AND p.code IN ('doc:create', 'doc:update', 'doc:publish')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (id, role_id, permission_id, data_scope)
SELECT uuid_generate_v4(), r.id, p.id, 'department'
FROM roles r
CROSS JOIN permissions p
WHERE r.code = 'vice_minister'
  AND p.code IN ('doc:create', 'doc:update')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO knowledge_categories (id, parent_id, name, slug, sort_order)
VALUES
    ('11111111-1111-4111-8111-111111111111', NULL, '协会公开', 'association', 10),
    ('22222222-2222-4222-8222-222222222222', NULL, '成员手册', 'handbook', 20)
ON CONFLICT DO NOTHING;

INSERT INTO knowledge_docs (
    id, kind, slug, title, summary, content, category_id, visibility, status, version, author_id, published_at
)
SELECT
    'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1',
    'page',
    'about-us',
    '关于我们',
    '深圳北理莫斯科大学计算机协会（StarByte）',
    $md$
# 关于 StarByte

深圳北理莫斯科大学计算机协会（StarByte）是面向全校的学生技术社团。我们一起学习、创造、协作，把想法做成作品。

## 我们做什么

- 技术分享与项目实践
- 协会一体化管理平台的建设与运营
- 招新、活动与成员成长

## 加入我们

登录后即可提交入会申请。公开资料：

- [计算机协会章程](/docs/association-charter)
- [使用手册](/docs/user-manual)（需登录）
- [API 手册说明](/docs/api-manual)（需登录）
$md$,
    '11111111-1111-4111-8111-111111111111',
    'public',
    1,
    1,
    u.id,
    CURRENT_TIMESTAMP
FROM users u
WHERE u.username = 'admin'
  AND NOT EXISTS (SELECT 1 FROM knowledge_docs d WHERE d.slug = 'about-us' AND d.deleted_at IS NULL)
LIMIT 1;

INSERT INTO knowledge_docs (
    id, kind, slug, title, summary, content, category_id, visibility, status, version, author_id, published_at
)
SELECT
    'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
    'doc',
    'association-charter',
    '计算机协会章程',
    '协会组织与运行的基本规则',
    $md$
# 计算机协会章程

本文为招新与日常查阅用的公开章程摘要。完整条文以协会正式文本为准。

## 第一章 总则

1. 本协会全称为深圳北理莫斯科大学计算机协会，对外可用 StarByte。
2. 协会是学生自愿组成的技术社团，接受学校相关指导。

## 第二章 宗旨

促进计算机科学学习与实践，培养协作与工程能力，服务校园信息化。

## 第三章 成员

- 在校学生可申请加入。
- 成员应遵守协会纪律与平台使用规范。

## 第四章 组织

协会设社长、副社长及各部门。日常协作通过 StarByte 平台进行。
$md$,
    '11111111-1111-4111-8111-111111111111',
    'public',
    1,
    1,
    u.id,
    CURRENT_TIMESTAMP
FROM users u
WHERE u.username = 'admin'
  AND NOT EXISTS (SELECT 1 FROM knowledge_docs d WHERE d.slug = 'association-charter' AND d.deleted_at IS NULL)
LIMIT 1;

INSERT INTO knowledge_docs (
    id, kind, slug, title, summary, content, category_id, visibility, status, version, author_id, published_at
)
SELECT
    'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
    'doc',
    'user-manual',
    '使用手册',
    '登录后的工作台与常用功能说明',
    $md$
# 使用手册

本手册面向已登录成员。招新同学请先完成注册或校园统一认证。

## 登录

- 首页 `/` 为登录门脸。
- 校园网建议使用学校统一认证。

## 工作台

登录后进入工作台，可处理入会申请、通知、任务与公告。

## 公告与文档

- **公告**是时效通知，见工作台「公告中心」。
- **文档 / 手册**在知识库中维护：公开页如 [关于我们](/about-us)、[章程](/docs/association-charter)。
$md$,
    '22222222-2222-4222-8222-222222222222',
    'authenticated',
    1,
    1,
    u.id,
    CURRENT_TIMESTAMP
FROM users u
WHERE u.username = 'admin'
  AND NOT EXISTS (SELECT 1 FROM knowledge_docs d WHERE d.slug = 'user-manual' AND d.deleted_at IS NULL)
LIMIT 1;

INSERT INTO knowledge_docs (
    id, kind, slug, title, summary, content, category_id, visibility, status, version, author_id, published_at
)
SELECT
    'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4',
    'doc',
    'api-manual',
    'API 手册说明',
    '如何获取令牌并打开 Swagger，不在此重复 OpenAPI',
    $md$
# API 手册说明

本页只说明如何调用 StarByte API，**不复制** OpenAPI 定义。接口契约以 Swagger UI 为准。

## 打开 Swagger

非生产环境访问：

- [Swagger UI](/swagger/index.html)

生产环境默认关闭 Swagger。

## 获取访问令牌

1. 使用 `/api/v1/auth/login` 或校园 CAS 换取 `access_token`。
2. 在请求头携带：`Authorization: Bearer <access_token>`。
3. 令牌过期后使用 `/api/v1/auth/refresh` 刷新。

## 权限

多数写接口需要对应 RBAC 权限码（如 `doc:read`）。超级管理员绕过校验。
$md$,
    '22222222-2222-4222-8222-222222222222',
    'authenticated',
    1,
    1,
    u.id,
    CURRENT_TIMESTAMP
FROM users u
WHERE u.username = 'admin'
  AND NOT EXISTS (SELECT 1 FROM knowledge_docs d WHERE d.slug = 'api-manual' AND d.deleted_at IS NULL)
LIMIT 1;

INSERT INTO knowledge_doc_versions (id, doc_id, version, title, summary, content, editor_id)
SELECT uuid_generate_v4(), d.id, d.version, d.title, d.summary, d.content, d.author_id
FROM knowledge_docs d
WHERE d.id IN (
    'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1',
    'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
    'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3',
    'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4'
)
AND NOT EXISTS (
    SELECT 1 FROM knowledge_doc_versions v WHERE v.doc_id = d.id AND v.version = d.version
);
