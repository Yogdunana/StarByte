package service

import "github.com/Yogdunana/StarByte/backend/pkg/response"

func notFound() error {
	return response.NewError(response.CodeKnowledgeNotFound, "文档不存在")
}

func categoryNotFound() error {
	return response.NewError(response.CodeKnowledgeNotFound, "分类不存在")
}

func invalidState(msg string) error {
	return response.NewError(response.CodeKnowledgeInvalidState, msg)
}

func noAccess(msg string) error {
	return response.NewError(response.CodeKnowledgeNoAccess, msg)
}

func loginRequired() error {
	return response.NewError(response.CodeKnowledgeLoginRequired, "阅读该文档需要登录")
}

func invalidSlug() error {
	return response.NewError(response.CodeKnowledgeInvalidSlug, "文档 slug 不合法")
}

func slugConflict() error {
	return response.NewError(response.CodeKnowledgeConflict, "文档 slug 已存在")
}

func invalidVis() error {
	return response.NewError(response.CodeKnowledgeInvalidVis, "文档可见性不合法")
}
