-- 公开页改走前端章程 HTML 与 About 落地页；CMS 摘要指向正式文本。
UPDATE knowledge_docs
SET
    summary = '深圳北理莫斯科大学计算机协会章程全文',
    content = $md$# 计算机协会章程

公开页 [/docs/association-charter](/docs/association-charter) 展示章程正式 HTML。

源文件：`docs/charter/smbu-ca-charter.html`。角色口径与一期未落实机构见 Issue #199、#200。
$md$,
    updated_at = CURRENT_TIMESTAMP
WHERE slug = 'association-charter'
  AND deleted_at IS NULL;

UPDATE knowledge_docs
SET
    summary = '深圳北理莫斯科大学计算机协会（深北莫计协 / SMBU-CA）',
    content = $md$# 关于我们

公开介绍页：[/about-us](/about-us)。章程全文：[/docs/association-charter](/docs/association-charter)。
$md$,
    updated_at = CURRENT_TIMESTAMP
WHERE slug = 'about-us'
  AND deleted_at IS NULL;
