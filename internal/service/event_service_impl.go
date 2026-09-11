// Package service 业务逻辑层实现 - 日程事件
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
//
// 本文件实现 service.ScheduleEventService 接口
// 编排: repo + TransactionManager + RecurrenceExpander
// 职责: 参数校验 → 权限校验 → 事务编排 → 错误映射
package service

import (
	"context"
	"errors"
	"time"

	"schedule-service/internal/model"
	"schedule-service/internal/repo"
)

// ============================================================================
// 实现 struct 与构造函数
// ============================================================================

// scheduleEventService 日程事件业务实现
type scheduleEventService struct {
	eventRepo    repo.ScheduleEventRepository
	attendeeRepo repo.ScheduleEventAttendeeRepository
	reminderRepo repo.ScheduleReminderRepository
	txMgr        repo.TransactionManager
	expander     RecurrenceExpander
}

// NewScheduleEventService 构造日程事件业务实现
func NewScheduleEventService(
	eventRepo repo.ScheduleEventRepository,
	attendeeRepo repo.ScheduleEventAttendeeRepository,
	reminderRepo repo.ScheduleReminderRepository,
	txMgr repo.TransactionManager,
	expander RecurrenceExpander,
) ScheduleEventService {
	return &scheduleEventService{
		eventRepo:    eventRepo,
		attendeeRepo: attendeeRepo,
		reminderRepo: reminderRepo,
		txMgr:        txMgr,
		expander:     expander,
	}
}

// ----------------------------------------------------------------------------
// 基础 CRUD
// ----------------------------------------------------------------------------

// Create 创建日程事件
//   事务内: 写 event → 写 attendees → 写 reminders
func (s *scheduleEventService) Create(ctx context.Context, actorID uint, req *CreateEventRequest) (*EventResponse, error) {
	// 1. 参数校验
	if req.Title == "" {
		return nil, NewBizError(ErrCodeInvalidParam, "title is required", nil)
	}
	if err := validateEventTime(req.StartTime, req.EndTime, req.AllDay); err != nil {
		return nil, err
	}
	if req.CalendarType == "" {
		req.CalendarType = model.CalendarTypePersonal
	}
	if !validateCalendarType(req.CalendarType) {
		return nil, NewBizError(ErrCodeInvalidParam, "invalid calendar_type", nil)
	}
	if req.ShareScope == "" {
		req.ShareScope = model.ShareScopePrivate
	}
	if !validateShareScope(req.ShareScope) {
		return nil, NewBizError(ErrCodeInvalidParam, "invalid share_scope", nil)
	}
	if req.Status == "" {
		req.Status = model.EventStatusDraft
	}
	if !validateEventStatus(req.Status) {
		return nil, NewBizError(ErrCodeInvalidParam, "invalid status", nil)
	}
	// RRULE 校验
	if req.RecurrenceRule != "" {
		if err := s.expander.Validate(ctx, req.RecurrenceRule); err != nil {
			return nil, err
		}
	}
	// 参与人角色校验
	for _, a := range req.Attendees {
		if !validateAttendeeRole(a.Role) {
			return nil, NewBizError(ErrCodeInvalidParam, "invalid attendee role", nil)
		}
	}
	// 提醒参数校验
	for _, r := range req.Reminders {
		if !validateRemindMethod(r.RemindMethod) {
			return nil, NewBizError(ErrCodeReminderMethodUnsupported, "unsupported remind_method", nil)
		}
		if r.RemindOffset < 0 || r.RemindOffset > maxReminderOffset {
			return nil, NewBizError(ErrCodeReminderOffsetInvalid, "remind_offset out of range", nil)
		}
	}

	// 2. 构造 event 实体
	event := &model.ScheduleEvent{
		Title:           req.Title,
		Description:     req.Description,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		AllDay:          req.AllDay,
		Location:        req.Location,
		CalendarType:    req.CalendarType,
		OwnerID:         actorID,
		ShareScope:      req.ShareScope,
		ShareTargetIDs:  formatShareTargetIDs(req.ShareTargetIDs),
		RecurrenceRule:  req.RecurrenceRule,
		Status:          req.Status,
		MeetingID:       req.MeetingID,
		CreatorID:       actorID,
		UpdaterID:       actorID,
		Version:         1,
	}

	// 3. 事务内: 写 event → 写 attendees → 写 reminders
	var createdEvent *model.ScheduleEvent
	err := s.txMgr.Transaction(ctx, func(txCtx context.Context) error {
		// 3.1 写 event
		if err := s.eventRepo.Create(txCtx, event); err != nil {
			return wrapEventErr(err)
		}
		createdEvent = event

		// 3.2 批量写 attendees（如果提供）
		if len(req.Attendees) > 0 {
			attendees := make([]*model.ScheduleEventAttendee, 0, len(req.Attendees))
			seen := make(map[uint]bool)
			for _, a := range req.Attendees {
				// 去重
				if seen[a.UserID] {
					continue
				}
				seen[a.UserID] = true
				attendees = append(attendees, &model.ScheduleEventAttendee{
					EventID:         event.ID,
					UserID:          a.UserID,
					Role:            a.Role,
					ResponseStatus:  model.AttendeeResponsePending,
				})
			}
			if err := s.attendeeRepo.BatchCreate(txCtx, attendees); err != nil {
				return wrapAttendeeErr(err)
			}
		}

		// 3.3 批量写 reminders（如果提供）
		for _, r := range req.Reminders {
			reminder := &model.ScheduleReminder{
				EventID:      event.ID,
				UserID:       r.UserID,
				RemindOffset: r.RemindOffset,
				RemindMethod: r.RemindMethod,
				Status:       model.ReminderStatusPending,
				RemindTime:   event.StartTime.Add(-time.Duration(r.RemindOffset) * time.Minute),
			}
			if err := s.reminderRepo.Create(txCtx, reminder); err != nil {
				return wrapReminderErr(err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// 4. 返回响应
	return eventToResponse(createdEvent), nil
}

// Update 更新日程事件
//   校验: 操作者是否为 owner、乐观锁版本号、状态转换合法性
func (s *scheduleEventService) Update(ctx context.Context, actorID, eventID uint, req *UpdateEventRequest) (*EventResponse, error) {
	// 1. 查询现有事件
	existing, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, wrapEventErr(err)
	}

	// 2. 权限校验: 仅 owner 可修改
	if existing.OwnerID != actorID {
		return nil, ErrForbidden
	}

	// 3. 参数校验
	if req.Title == "" {
		return nil, NewBizError(ErrCodeInvalidParam, "title is required", nil)
	}
	if err := validateEventTime(req.StartTime, req.EndTime, req.AllDay); err != nil {
		return nil, err
	}
	if !validateCalendarType(req.CalendarType) {
		return nil, NewBizError(ErrCodeInvalidParam, "invalid calendar_type", nil)
	}
	if !validateShareScope(req.ShareScope) {
		return nil, NewBizError(ErrCodeInvalidParam, "invalid share_scope", nil)
	}
	if !validateEventStatus(req.Status) {
		return nil, NewBizError(ErrCodeInvalidParam, "invalid status", nil)
	}
	if req.RecurrenceRule != "" {
		if err := s.expander.Validate(ctx, req.RecurrenceRule); err != nil {
			return nil, err
		}
	}

	// 4. 状态转换校验
	if !eventStatusTransitionAllowed(existing.Status, req.Status) {
		return nil, NewBizError(ErrCodeEventStatusTransition, "illegal status transition: "+existing.Status+" → "+req.Status, nil)
	}

	// 5. 构造更新实体（携带原 Version 用于乐观锁）
	event := &model.ScheduleEvent{
		BaseModel: model.BaseModel{ID: existing.ID},
		Title:           req.Title,
		Description:     req.Description,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		AllDay:          req.AllDay,
		Location:        req.Location,
		CalendarType:    req.CalendarType,
		ShareScope:      req.ShareScope,
		ShareTargetIDs:  formatShareTargetIDs(req.ShareTargetIDs),
		RecurrenceRule:  req.RecurrenceRule,
		Status:          req.Status,
		MeetingID:       req.MeetingID,
		UpdaterID:       actorID,
		Version:         req.Version,
	}

	if err := s.eventRepo.Update(ctx, event); err != nil {
		if errors.Is(err, repo.ErrConcurrent) {
			return nil, NewBizError(ErrCodeEventConcurrent, "event has been modified by others, please refresh", err)
		}
		return nil, wrapEventErr(err)
	}

	return eventToResponse(event), nil
}

// Delete 删除日程事件
//   事务内: 软删 event → 取消 reminders
//   attendees 通过事件软删后自动失效（查询都 JOIN 事件）
func (s *scheduleEventService) Delete(ctx context.Context, actorID, eventID uint) error {
	// 1. 权限校验
	existing, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return wrapEventErr(err)
	}
	if existing.OwnerID != actorID {
		return ErrForbidden
	}

	// 2. 事务内删除
	return s.txMgr.Transaction(ctx, func(txCtx context.Context) error {
		// 2.1 软删事件
		if err := s.eventRepo.Delete(txCtx, eventID); err != nil {
			return wrapEventErr(err)
		}
		// 2.2 批量取消提醒
		if err := s.reminderRepo.CancelByEvent(txCtx, eventID); err != nil {
			return wrapReminderErr(err)
		}
		return nil
	})
}

// GetByID 查询单个事件详情
//   权限校验: 操作者须为 owner 或被共享者或参与人
//   填充 attendees 与当前用户的 reminders
func (s *scheduleEventService) GetByID(ctx context.Context, actorID, eventID uint) (*EventResponse, error) {
	// 1. 查询事件
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, wrapEventErr(err)
	}

	// 2. 权限校验
	if !s.canAccessEvent(ctx, event, actorID) {
		return nil, NewBizError(ErrCodeForbidden, "no permission to access this event", nil)
	}

	// 3. 转换响应
	resp := eventToResponse(event)

	// 4. 填充参与人列表（仅 owner 可看完整列表）
	if event.OwnerID == actorID {
		attendees, err := s.attendeeRepo.ListByEvent(ctx, eventID)
		if err != nil {
			// 列表查询失败不阻塞主流程，仅日志记录
		} else {
			resp.Attendees = make([]*AttendeeResponse, 0, len(attendees))
			for _, a := range attendees {
				resp.Attendees = append(resp.Attendees, attendeeToResponse(a))
			}
		}
	}

	// 5. 填充当前用户的提醒列表
	reminders, err := s.reminderRepo.ListByEvent(ctx, eventID)
	if err == nil {
		resp.Reminders = make([]*ReminderResponse, 0, len(reminders))
		for _, r := range reminders {
			if r.UserID == actorID {
				resp.Reminders = append(resp.Reminders, reminderToResponse(r))
			}
		}
	}

	return resp, nil
}

// ----------------------------------------------------------------------------
// 列表查询
// ----------------------------------------------------------------------------

// List 组合条件分页查询
func (s *scheduleEventService) List(ctx context.Context, actorID uint, req *EventListRequest) (*EventListResponse, error) {
	// 默认查询当前用户拥有的事件
	opts := repo.EventQueryOptions{
		OwnerID:          req.OwnerID,
		CalendarType:     req.CalendarType,
		Status:           req.Status,
		StartTimeFrom:    req.StartTimeFrom,
		StartTimeTo:      req.StartTimeTo,
		MeetingID:        req.MeetingID,
		ExcludeCancelled: req.ExcludeCancelled,
	}
	// 如果未指定 ownerID，默认查当前用户的
	if opts.OwnerID == 0 {
		opts.OwnerID = actorID
	}
	params := toRepoQueryParams(req.PageRequest)
	list, err := s.eventRepo.List(ctx, opts, params)
	if err != nil {
		return nil, wrapEventErr(err)
	}
	return eventsToListResponse(list.List, list.Total, list.Page, list.PageSize), nil
}

// ListByTimeRange 日/周/月视图查询
//   返回: 普通事件 + 重复事件展开后的虚拟实例
func (s *scheduleEventService) ListByTimeRange(ctx context.Context, actorID uint, from, to time.Time) ([]*EventResponse, error) {
	// 1. 查询母事件
	events, err := s.eventRepo.ListByTimeRange(ctx, actorID, from, to)
	if err != nil {
		return nil, wrapEventErr(err)
	}

	result := make([]*EventResponse, 0, len(events))
	for _, e := range events {
		// 2. 非重复事件直接加入
		if e.RecurrenceRule == "" {
			// 仅返回区间内有交集的事件（repo 已过滤）
			result = append(result, eventToResponse(e))
			continue
		}

		// 3. 重复事件：展开 RRULE 生成虚拟实例
		//    首先加入母事件本身（如果在区间内）
		if !e.StartTime.Before(from) && !e.StartTime.After(to) {
			result = append(result, eventToResponse(e))
		}
		// 展开实例
		instances, err := s.expander.Expand(ctx, e.RecurrenceRule, e.StartTime, from, to)
		if err != nil {
			// 展开失败不阻塞，跳过此事件的展开
			continue
		}
		// 生成虚拟响应（ID=0 标识为虚拟实例）
		for _, inst := range instances {
			if inst.Equal(e.StartTime) {
				continue // 跳过与母事件重复的
			}
			resp := eventToResponse(e)
			resp.ID = 0          // 虚拟实例无真实 ID
			resp.StartTime = inst
			resp.EndTime = inst.Add(e.EndTime.Sub(e.StartTime))
			result = append(result, resp)
		}
	}

	return result, nil
}

// ListVisibleByUser 查询用户可见的所有日程（自己拥有 + 被共享）
func (s *scheduleEventService) ListVisibleByUser(ctx context.Context, actorID uint, req *PageRequest) (*EventListResponse, error) {
	params := toRepoQueryParams(*req)
	list, err := s.eventRepo.ListVisibleByUser(ctx, actorID, params)
	if err != nil {
		return nil, wrapEventErr(err)
	}
	return eventsToListResponse(list.List, list.Total, list.Page, list.PageSize), nil
}

// ListByAttendee 查询用户作为参与人的日程
func (s *scheduleEventService) ListByAttendee(ctx context.Context, actorID uint, req *PageRequest) (*EventListResponse, error) {
	params := toRepoQueryParams(*req)
	list, err := s.eventRepo.ListByAttendee(ctx, actorID, params)
	if err != nil {
		return nil, wrapEventErr(err)
	}
	return eventsToListResponse(list.List, list.Total, list.Page, list.PageSize), nil
}

// ----------------------------------------------------------------------------
// 状态与关联
// ----------------------------------------------------------------------------

// UpdateStatus 更新事件状态
func (s *scheduleEventService) UpdateStatus(ctx context.Context, actorID, eventID uint, status string) error {
	// 1. 权限校验
	existing, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return wrapEventErr(err)
	}
	if existing.OwnerID != actorID {
		return ErrForbidden
	}
	// 2. 状态校验
	if !validateEventStatus(status) {
		return NewBizError(ErrCodeInvalidParam, "invalid status", nil)
	}
	// 3. 状态转换校验
	if !eventStatusTransitionAllowed(existing.Status, status) {
		return NewBizError(ErrCodeEventStatusTransition, "illegal status transition", nil)
	}
	// 4. 更新
	if err := s.eventRepo.UpdateStatus(ctx, eventID, status); err != nil {
		return wrapEventErr(err)
	}
	return nil
}

// LinkMeeting 关联会议
func (s *scheduleEventService) LinkMeeting(ctx context.Context, actorID, eventID, meetingID uint) error {
	// 1. 权限校验
	existing, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return wrapEventErr(err)
	}
	if existing.OwnerID != actorID {
		return ErrForbidden
	}
	// 2. 防重复关联
	if existing.MeetingID != nil {
		return NewBizError(ErrCodeEventMeetingLinked, "event already linked to a meeting", nil)
	}
	// 3. 关联
	if err := s.eventRepo.UpdateMeetingID(ctx, eventID, &meetingID); err != nil {
		return wrapEventErr(err)
	}
	return nil
}

// UnlinkMeeting 取消会议关联
func (s *scheduleEventService) UnlinkMeeting(ctx context.Context, actorID, eventID uint) error {
	// 1. 权限校验
	existing, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return wrapEventErr(err)
	}
	if existing.OwnerID != actorID {
		return ErrForbidden
	}
	// 2. 取消关联
	if err := s.eventRepo.UpdateMeetingID(ctx, eventID, nil); err != nil {
		return wrapEventErr(err)
	}
	return nil
}

// ----------------------------------------------------------------------------
// 内部辅助
// ----------------------------------------------------------------------------

// canAccessEvent 判断用户是否有权访问某事件
//   规则: owner 可访问自己的 + 被共享可访问 + 作为参与人可访问
func (s *scheduleEventService) canAccessEvent(ctx context.Context, event *model.ScheduleEvent, userID uint) bool {
	// 1. owner 直接通过
	if event.OwnerID == userID {
		return true
	}
	// 2. 共享事件: 检查用户是否在被共享列表中
	if event.CalendarType == model.CalendarTypeShared {
		switch event.ShareScope {
		case model.ShareScopePublic:
			return true
		case model.ShareScopeUser:
			// 检查 userID 是否在 ShareTargetIDs 中
			targets := parseShareTargetIDs(event.ShareTargetIDs)
			for _, t := range targets {
				if t == userID {
					return true
				}
			}
		}
	}
	// 3. 参与人: 通过 attendeeRepo 查询
	_, err := s.attendeeRepo.GetByEventAndUser(ctx, event.ID, userID)
	return err == nil
}
