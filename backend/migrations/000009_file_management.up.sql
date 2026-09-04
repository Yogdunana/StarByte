-- ============================================================
-- 000009_file_management.up.sql
-- 文件管理：分类/公开/缩略图字段 + 权限码
-- 对齐现有 files 表列名（name/path/size/bucket/uploaded_by），不重命名
-- ============================================================

ALTER TABLE files ADD COLUMN IF NOT EXISTS category VARCHAR(20) DEFAULT 'document';
ALTER TABLE files ADD COLUMN IF NOT EXISTS is_public BOOLEAN DEFAULT FALSE;
ALTER TABLE files ADD COLUMN IF NOT EXISTS thumbnail_path VARCHAR(500);

CREATE INDEX IF NOT EXISTS idx_files_category ON files(category);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '文件查看', 'file:read', 'file', 'read', '查询文件列表与详情、下载', 3, true, 0),
    (uuid_generate_v4(), '文件上传', 'file:create', 'file', 'create', '上传单个或批量文件', 2, true, 0),
    (uuid_generate_v4(), '文件删除', 'file:delete', 'file', 'delete', '删除任意文件（上传者无需此权限）', 2, true, 0)
ON CONFLICT (code) DO NOTHING;
