package service

import "github.com/Yogdunana/StarByte/backend/pkg/search"

func fts(parts ...string) string {
	expr := "starbyte_cjk_tokens("
	for i, p := range parts {
		if i > 0 {
			expr += " || ' ' || "
		}
		expr += "coalesce(" + p + ",'')"
	}
	return expr + ")"
}

func headline(parts ...string) string {
	expr := ""
	for i, p := range parts {
		if i > 0 {
			expr += " || ' ' || "
		}
		expr += "coalesce(" + p + ",'')"
	}
	return expr
}

func catalogs() []search.Schema {
	return []search.Schema{
		{
			Code: "users", Name: "用户", Table: "users", IDColumn: "id",
			FTSExpr:      fts("username", "real_name", "email", "phone"),
			Headline:     headline("username", "real_name", "email"),
			ExtraWhere:   "deleted_at IS NULL",
			RBACResource: "user",
			ScopeColumn:  "department_id",
			SelfSQL:      `t."id" = ?`,
			Fields: []search.Field{
				{Name: "id", Column: "id", Kind: search.KindString, Label: "ID", Sortable: true, Filterable: true},
				{Name: "username", Column: "username", Kind: search.KindString, Label: "用户名", Searchable: true, Filterable: true, Sortable: true},
				{Name: "real_name", Column: "real_name", Kind: search.KindString, Label: "姓名", Searchable: true, Filterable: true, Sortable: true},
				{Name: "email", Column: "email", Kind: search.KindString, Label: "邮箱", Searchable: true, Filterable: true},
				{Name: "phone", Column: "phone", Kind: search.KindString, Label: "手机", Searchable: true, Filterable: true},
				{Name: "status", Column: "status", Kind: search.KindNumber, Label: "状态", Filterable: true, Sortable: true, Agg: true},
				{Name: "created_at", Column: "created_at", Kind: search.KindTime, Label: "创建时间", Filterable: true, Sortable: true, Agg: true},
			},
		},
		{
			Code: "tasks", Name: "任务", Table: "tasks", IDColumn: "id",
			FTSExpr:      fts("title", "description", "tags"),
			Headline:     headline("title", "description"),
			RBACResource: "task:read",
			ExtraWhere:   "deleted_at IS NULL",
			ScopeColumn:  "department_id",
			SelfSQL:      `t."creator_id" = ? OR t."assignee_id" = ?`,
			Fields: []search.Field{
				{Name: "id", Column: "id", Kind: search.KindString, Label: "ID", Sortable: true, Filterable: true},
				{Name: "title", Column: "title", Kind: search.KindString, Label: "标题", Searchable: true, Filterable: true, Sortable: true},
				{Name: "description", Column: "description", Kind: search.KindString, Label: "描述", Searchable: true, Filterable: true},
				{Name: "status", Column: "status", Kind: search.KindNumber, Label: "状态", Filterable: true, Sortable: true, Agg: true},
				{Name: "priority", Column: "priority", Kind: search.KindNumber, Label: "优先级", Filterable: true, Sortable: true, Agg: true},
				{Name: "progress", Column: "progress", Kind: search.KindNumber, Label: "进度", Filterable: true, Sortable: true, Agg: true},
				{Name: "tags", Column: "tags", Kind: search.KindString, Label: "标签", Searchable: true, Filterable: true},
				{Name: "created_at", Column: "created_at", Kind: search.KindTime, Label: "创建时间", Filterable: true, Sortable: true, Agg: true},
			},
		},
		{
			Code: "audit_logs", Name: "审计日志", Table: "audit_logs", IDColumn: "id",
			FTSExpr:      fts("operation", "path", "username", "real_name", "module"),
			Headline:     headline("operation", "path", "username"),
			RBACResource: "audit",
			Fields: []search.Field{
				{Name: "id", Column: "id", Kind: search.KindString, Label: "ID", Sortable: true, Filterable: true},
				{Name: "username", Column: "username", Kind: search.KindString, Label: "用户名", Searchable: true, Filterable: true, Sortable: true},
				{Name: "real_name", Column: "real_name", Kind: search.KindString, Label: "姓名", Searchable: true, Filterable: true},
				{Name: "operation", Column: "operation", Kind: search.KindString, Label: "操作", Searchable: true, Filterable: true, Sortable: true},
				{Name: "module", Column: "module", Kind: search.KindString, Label: "模块", Searchable: true, Filterable: true, Sortable: true, Agg: true},
				{Name: "action", Column: "action", Kind: search.KindString, Label: "动作", Filterable: true, Sortable: true, Agg: true},
				{Name: "path", Column: "path", Kind: search.KindString, Label: "路径", Searchable: true, Filterable: true},
				{Name: "method", Column: "method", Kind: search.KindString, Label: "方法", Filterable: true, Agg: true},
				{Name: "response_status", Column: "response_status", Kind: search.KindNumber, Label: "状态码", Filterable: true, Agg: true},
				{Name: "duration_ms", Column: "duration_ms", Kind: search.KindNumber, Label: "耗时(ms)", Filterable: true, Agg: true},
				{Name: "created_at", Column: "created_at", Kind: search.KindTime, Label: "时间", Filterable: true, Sortable: true, Agg: true},
			},
		},
		{
			Code: "member_applications", Name: "入会申请", Table: "member_applications", IDColumn: "id",
			FTSExpr:      fts("reason", "current_stage"),
			Headline:     headline("reason", "current_stage"),
			RBACResource: "member",
			ScopeColumn:  "department_id",
			SelfSQL:      `t."user_id" = ?`,
			Fields: []search.Field{
				{Name: "id", Column: "id", Kind: search.KindString, Label: "ID", Sortable: true, Filterable: true},
				{Name: "type", Column: "type", Kind: search.KindNumber, Label: "类型", Filterable: true, Sortable: true, Agg: true},
				{Name: "status", Column: "status", Kind: search.KindNumber, Label: "状态", Filterable: true, Sortable: true, Agg: true},
				{Name: "reason", Column: "reason", Kind: search.KindString, Label: "申请理由", Searchable: true, Filterable: true},
				{Name: "current_stage", Column: "current_stage", Kind: search.KindString, Label: "当前阶段", Searchable: true, Filterable: true, Sortable: true},
				{Name: "submitted_at", Column: "submitted_at", Kind: search.KindTime, Label: "提交时间", Filterable: true, Sortable: true, Agg: true},
				{Name: "created_at", Column: "created_at", Kind: search.KindTime, Label: "创建时间", Filterable: true, Sortable: true, Agg: true},
			},
		},
	}
}
