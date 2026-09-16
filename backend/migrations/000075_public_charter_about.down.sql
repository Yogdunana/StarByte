UPDATE knowledge_docs
SET
    summary = '协会组织与运行的基本规则',
    content = $md$# 计算机协会章程

本文为招新与日常查阅用的公开章程摘要。完整条文以协会正式文本为准。
$md$,
    updated_at = CURRENT_TIMESTAMP
WHERE slug = 'association-charter'
  AND deleted_at IS NULL;

UPDATE knowledge_docs
SET
    summary = '深圳北理莫斯科大学计算机协会（StarByte）',
    content = $md$# 关于 StarByte

深圳北理莫斯科大学计算机协会（StarByte）是面向全校的学生技术社团。
$md$,
    updated_at = CURRENT_TIMESTAMP
WHERE slug = 'about-us'
  AND deleted_at IS NULL;
