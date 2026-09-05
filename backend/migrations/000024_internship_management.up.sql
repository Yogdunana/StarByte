-- ============================================================
-- 000024_internship_management.up.sql
-- Issue #10：复用 000014 internships，不改已有列名
-- status 对齐 Issue：0进行中 1已完成 2已中止
--   （覆盖 000014 注释：1已结束 2已取消；表尚无业务数据）
-- internship_records 保留给打卡扩展，本期不使用
-- ============================================================

ALTER TABLE internships
    ADD COLUMN IF NOT EXISTS title VARCHAR(200) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS organization VARCHAR(200) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS type SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS mentor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS supervisor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS skills TEXT NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS achievements TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS report TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS mentor_comment TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS updated_by UUID REFERENCES users(id) ON DELETE SET NULL;

UPDATE internships
SET created_by = user_id
WHERE created_by IS NULL;

CREATE INDEX IF NOT EXISTS idx_internships_type ON internships(type);
CREATE INDEX IF NOT EXISTS idx_internships_mentor_id ON internships(mentor_id);
CREATE INDEX IF NOT EXISTS idx_internships_start_date ON internships(start_date);

INSERT INTO permissions (id, name, code, resource, action, description, type, is_system, status)
VALUES
    (uuid_generate_v4(), '实习评价', 'internship:evaluate', 'internship', 'evaluate', '指导老师评价实习', 2, true, 0)
ON CONFLICT (code) DO NOTHING;

INSERT INTO configs (id, config_key, config_value, config_type, description, category, is_public)
VALUES (
    uuid_generate_v4(),
    'internship_config',
    '{"allow_student_edit":true,"allow_minister_edit":true,"ranking_visible":true}',
    'json',
    '实习权限开关：学生可改/部长可改/排行榜可见',
    'internship',
    false
)
ON CONFLICT (config_key) DO NOTHING;
