package service

import (
	"context"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Yogdunana/StarByte/backend/internal/leave/dto"
	"github.com/Yogdunana/StarByte/backend/internal/leave/model"
	"github.com/google/uuid"
)

var typeCodeRe = regexp.MustCompile(`^[a-z][a-z0-9_]{0,19}$`)

func (s *leaveService) CreateType(ctx context.Context, viewer Viewer, req *dto.UpsertLeaveTypeRequest) (*dto.LeaveTypeResponse, error) {
	if !viewer.CanApprove {
		return nil, noAccess("无权配置请假类型")
	}
	code, name, err := normalizeType(req.Code, req.Name)
	if err != nil {
		return nil, err
	}
	existing, err := s.rows.GetLeaveTypeByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, typeInvalid("请假类型编码已存在")
	}
	now := s.clock()
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	row := &model.LeaveType{
		ID: uuid.New(), Name: name, Code: code, Deductible: req.Deductible,
		DefaultDays: req.DefaultDays, Description: strings.TrimSpace(req.Description),
		Enabled: enabled, SortOrder: req.SortOrder, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.rows.CreateLeaveType(ctx, row); err != nil {
		return nil, err
	}
	out := mapType(*row)
	return &out, nil
}

func (s *leaveService) UpdateType(ctx context.Context, viewer Viewer, id uuid.UUID, req *dto.UpsertLeaveTypeRequest) (*dto.LeaveTypeResponse, error) {
	if !viewer.CanApprove {
		return nil, noAccess("无权配置请假类型")
	}
	row, err := s.rows.GetLeaveTypeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, typeNotFound()
	}
	code, name, err := normalizeType(req.Code, req.Name)
	if err != nil {
		return nil, err
	}
	if other, err := s.rows.GetLeaveTypeByCode(ctx, code); err != nil {
		return nil, err
	} else if other != nil && other.ID != id {
		return nil, typeInvalid("请假类型编码已存在")
	}
	row.Name = name
	row.Code = code
	row.Deductible = req.Deductible
	row.DefaultDays = req.DefaultDays
	row.Description = strings.TrimSpace(req.Description)
	row.SortOrder = req.SortOrder
	if req.Enabled != nil {
		row.Enabled = *req.Enabled
	}
	row.UpdatedAt = s.clock()
	if err := s.rows.UpdateLeaveType(ctx, row); err != nil {
		return nil, err
	}
	out := mapType(*row)
	return &out, nil
}

func normalizeType(code, name string) (string, string, error) {
	code = strings.ToLower(strings.TrimSpace(code))
	name = strings.TrimSpace(name)
	if !typeCodeRe.MatchString(code) {
		return "", "", typeInvalid("类型编码须为小写字母开头的字母数字或下划线")
	}
	if name == "" || utf8.RuneCountInString(name) > model.MaxTypeNameLen {
		return "", "", typeInvalid("类型名称不合法")
	}
	return code, name, nil
}

func parseAttachments(items []dto.Attachment) (model.AttachmentList, error) {
	if len(items) > model.MaxAttachments {
		return nil, invalidTime("附件数量不能超过5个")
	}
	out := make(model.AttachmentList, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		id := strings.TrimSpace(item.FileID)
		if id == "" {
			continue
		}
		if _, err := uuid.Parse(id); err != nil {
			return nil, invalidTime("附件文件ID不合法")
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = id
		}
		if utf8.RuneCountInString(name) > 255 {
			name = string([]rune(name)[:255])
		}
		size := item.Size
		if size < 0 {
			size = 0
		}
		out = append(out, model.Attachment{FileID: id, Name: name, Size: size})
	}
	return out, nil
}

func parseCalendarRange(fromRaw, toRaw string, now time.Time) (time.Time, time.Time, error) {
	loc := bizLocation()
	if strings.TrimSpace(fromRaw) == "" && strings.TrimSpace(toRaw) == "" {
		start := time.Date(now.In(loc).Year(), now.In(loc).Month(), 1, 0, 0, 0, 0, loc)
		return start, start.AddDate(0, 1, 0), nil
	}
	from, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(fromRaw), loc)
	if err != nil {
		return time.Time{}, time.Time{}, invalidTime("日历起始日期格式错误")
	}
	to, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(toRaw), loc)
	if err != nil {
		return time.Time{}, time.Time{}, invalidTime("日历结束日期格式错误")
	}
	to = to.Add(24 * time.Hour)
	if !from.Before(to) {
		return time.Time{}, time.Time{}, invalidTime("日历起始日期必须早于结束日期")
	}
	if to.Sub(from) > 400*24*time.Hour {
		return time.Time{}, time.Time{}, invalidTime("日历跨度不能超过一年")
	}
	return from, to, nil
}
