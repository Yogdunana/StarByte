package service

import (
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/Yogdunana/StarByte/backend/internal/meeting/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
)

func validateMeeting(m *model.Meeting) error {
	if strings.TrimSpace(m.Title) == "" || utf8.RuneCountInString(m.Title) > 200 || m.MeetingType < 1 || m.MeetingType > 3 {
		return response.NewError(response.CodeBadRequest, "请填写有效的会议标题和类型")
	}
	if utf8.RuneCountInString(m.Location) > 200 || len(m.OnlineLink) > 500 {
		return response.NewError(response.CodeBadRequest, "会议地点或链接过长")
	}
	if m.OnlineLink != "" {
		link, err := url.Parse(m.OnlineLink)
		if err != nil || (link.Scheme != "http" && link.Scheme != "https") || link.Host == "" {
			return response.NewError(response.CodeBadRequest, "线上会议链接必须是完整的 http 或 https 地址")
		}
	}
	return nil
}
func validateAgenda(a *model.Agenda) error {
	if strings.TrimSpace(a.Title) == "" || utf8.RuneCountInString(a.Title) > 200 || utf8.RuneCountInString(a.Presenter) > 100 || (a.Duration != nil && *a.Duration < 0) {
		return response.NewError(response.CodeBadRequest, "议程标题不能为空，时长不能为负数，标题和汇报人不能超过长度限制")
	}
	return nil
}
