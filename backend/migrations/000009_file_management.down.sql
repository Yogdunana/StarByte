DELETE FROM permissions WHERE code IN ('file:read', 'file:create', 'file:delete');

DROP INDEX IF EXISTS idx_files_category;

ALTER TABLE files DROP COLUMN IF EXISTS thumbnail_path;
ALTER TABLE files DROP COLUMN IF EXISTS is_public;
ALTER TABLE files DROP COLUMN IF EXISTS category;
