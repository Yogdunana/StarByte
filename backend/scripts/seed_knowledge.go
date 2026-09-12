package main

import (
	"fmt"

	"gorm.io/gorm"
)

func seedKnowledge(db *gorm.DB) error {
	if err := db.Exec(`
		INSERT INTO knowledge_categories (id, parent_id, name, slug, sort_order)
		VALUES
			('11111111-1111-4111-8111-111111111111', NULL, '协会公开', 'association', 10),
			('22222222-2222-4222-8222-222222222222', NULL, '成员手册', 'handbook', 20)
		ON CONFLICT (id) DO NOTHING
	`).Error; err != nil {
		return fmt.Errorf("seed knowledge categories: %w", err)
	}

	type seedDoc struct {
		ID, Kind, Slug, Title, Summary, Content, CategoryID, Visibility string
	}
	docs := []seedDoc{
		{
			ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1", Kind: "page", Slug: "about-us",
			Title: "关于我们", Summary: "深圳北理莫斯科大学计算机协会（StarByte）",
			CategoryID: "11111111-1111-4111-8111-111111111111", Visibility: "public",
			Content: "# 关于 StarByte\n\n深圳北理莫斯科大学计算机协会（StarByte）是面向全校的学生技术社团。\n\n- [计算机协会章程](/docs/association-charter)\n- [使用手册](/docs/user-manual)\n- [API 手册说明](/docs/api-manual)\n",
		},
		{
			ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2", Kind: "doc", Slug: "association-charter",
			Title: "计算机协会章程", Summary: "协会组织与运行的基本规则",
			CategoryID: "11111111-1111-4111-8111-111111111111", Visibility: "public",
			Content: "# 计算机协会章程\n\n本文为招新与日常查阅用的公开章程摘要。\n\n## 宗旨\n\n促进计算机科学学习与实践，培养协作与工程能力。\n",
		},
		{
			ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3", Kind: "doc", Slug: "user-manual",
			Title: "使用手册", Summary: "登录后的工作台与常用功能说明",
			CategoryID: "22222222-2222-4222-8222-222222222222", Visibility: "authenticated",
			Content: "# 使用手册\n\n本手册面向已登录成员。首页 `/` 为登录门脸。\n",
		},
		{
			ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4", Kind: "doc", Slug: "api-manual",
			Title: "API 手册说明", Summary: "如何获取令牌并打开 Swagger，不在此重复 OpenAPI",
			CategoryID: "22222222-2222-4222-8222-222222222222", Visibility: "authenticated",
			Content: "# API 手册说明\n\n非生产环境打开 [Swagger UI](/swagger/index.html)。\n\n请求头：`Authorization: Bearer <access_token>`。\n",
		},
	}
	for _, d := range docs {
		if err := db.Exec(`
			INSERT INTO knowledge_docs (
				id, kind, slug, title, summary, content, category_id, visibility, status, version, author_id, published_at
			)
			SELECT ?, ?, ?, ?, ?, ?, ?, ?, 1, 1, u.id, CURRENT_TIMESTAMP
			FROM users u
			WHERE u.username = 'admin'
			  AND NOT EXISTS (SELECT 1 FROM knowledge_docs x WHERE x.slug = ? AND x.deleted_at IS NULL)
			LIMIT 1
		`, d.ID, d.Kind, d.Slug, d.Title, d.Summary, d.Content, d.CategoryID, d.Visibility, d.Slug).Error; err != nil {
			return fmt.Errorf("seed knowledge doc %s: %w", d.Slug, err)
		}
	}
	if err := db.Exec(`
		INSERT INTO knowledge_doc_versions (id, doc_id, version, title, summary, content, editor_id)
		SELECT uuid_generate_v4(), d.id, d.version, d.title, d.summary, d.content, d.author_id
		FROM knowledge_docs d
		WHERE d.slug IN ('about-us', 'association-charter', 'user-manual', 'api-manual')
		  AND NOT EXISTS (
			SELECT 1 FROM knowledge_doc_versions v WHERE v.doc_id = d.id AND v.version = d.version
		  )
	`).Error; err != nil {
		return fmt.Errorf("seed knowledge versions: %w", err)
	}
	return nil
}
