package service

import "github.com/Yogdunana/StarByte/backend/pkg/response"

func notFound() error {
	return response.NewError(response.CodeAnnouncementNotFound, "公告不存在")
}

func invalidState(msg string) error {
	return response.NewError(response.CodeAnnouncementInvalidState, msg)
}

func noAccess(msg string) error {
	return response.NewError(response.CodeAnnouncementNoAccess, msg)
}

func invalidCategory() error {
	return response.NewError(response.CodeAnnouncementInvalidCat, "公告分类不合法")
}
