// Package service 业务逻辑层实现 - 参与人
// 日程管理模块 (Schedule Module) - 对应 GitHub Issue #78
//
// 本文件实现 service.ScheduleAttendeeService 接口
// 编排: repo + TransactionManager
// 职责: 参数校验 → 权限校验 → 唯一性校验 → 错误映射
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

// scheduleAttendeeService 参与人业务实现
type scheduleAttendeeService struct {
	attendeeRepo repo.ScheduleEventAttendeeRepository
	eventRepo    repo.ScheduleEventRepository
	txMgr        repo.TransactionManager
}

// NewScheduleAttendeeService 构造参与人业务实现
func NewScheduleAttendeeService(
	attendeeRepo repo.ScheduleEventAttendeeRepository,
	eventRepo repo.ScheduleEventRepository,
	txMgr repo.TransactionManager,
) ScheduleAttendeeService {
	return &scheduleAttendeeService{
		attendeeRepo: attendeeRepo,
		eventRepo:    eventRepo,
		txMgr:        txMgr,
	}
}

// ----------------------------------------------------------------------------
// 基础操作
// ----------------------------------------------------------------------------

// Add 添加参与人到事件
//   校验: 事件存在、操作者为 owner、用户未已是参与人、数量未超限
func (s *scheduleAttendeeService) Add(ctx context.Context, actorID, eventID uint, req *CreateAttendeeRequest) (*AttendeeResponse, error) {
	// 1. 参数校验
	if req.UserID == 0 {
		return nil, NewBizError(ErrCodeInvalidParam, "user_id is required", nil)
	}
	if req.Role == "" {
		req.Role = model.AttendeeRoleRequired
	}
	if !validateAttendeeRole(req.Role) {
		return nil, NewBizError(ErrCodeInvalidParam, "invalid role", nil)
	}
	// 组织者角色只能由事件创建逻辑添加，外部不可直接指定
	if req.Role == model.AttendeeRoleOrganizer {
		return nil, NewBizError(ErrCodeInvalidParam, "cannot add organizer via API", nil)
	}

	// 2. 事件存在校验 + 权限校验
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, wrapEventErr(err)
	}
	if event.OwnerID != actorID {
		return nil, ErrForbidden
	}

	// 3. 唯一性校验: 用户未已是参与人
	existing, err := s.attendeeRepo.GetByEventAndUser(ctx, eventID, req.UserID)
	if err != nil && !errors.Is(err, repo.ErrNotFound) {
		return nil, wrapAttendeeErr(err)
	}
	if existing != nil {
		return nil, NewBizError(ErrCodeAttendeeDuplicate, "user is already an attendee", nil)
	}

	// 4. 数量上限校验
	count, err := s.attendeeRepo.CountByEvent(ctx, eventID, "")
	if err != nil {
		return nil, wrapAttendeeErr(err)
	}
	if count >= maxAttendeesPerEvent {
		return nil, NewBizError(ErrCodeAttendeeCountExceed, "attendee count exceeds limit", nil)
	}

	// 5. 创建
	attendee := &model.ScheduleEventAttendee{
		EventID:        eventID,
		UserID:         req.UserID,
		Role:           req.Role,
		ResponseStatus: model.AttendeeResponsePending,
	}
	if err := s.attendeeRepo.Create(ctx, attendee); err != nil {
		return nil, wrapAttendeeErr(err)
	}

	return attendeeToResponse(attendee), nil
}

// BatchAdd 批量添加参与人
//   事务内批量写入，去重 + 数量校验
func (s *scheduleAttendeeService) BatchAdd(ctx context.Context, actorID, eventID uint, reqs []CreateAttendeeRequest) ([]*AttendeeResponse, error) {
	// 1. 事件存在校验 + 权限校验
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, wrapEventErr(err)
	}
	if event.OwnerID != actorID {
		return nil, ErrForbidden
	}

	// 2. 参数校验 + 去重
	seen := make(map[uint]bool)
	var validReqs []*CreateAttendeeRequest
	for _, r := range reqs {
		if r.UserID == 0 {
			return nil, NewBizError(ErrCodeInvalidParam, "user_id is required", nil)
		}
		if r.Role == "" {
			r.Role = model.AttendeeRoleRequired
		}
		if !validateAttendeeRole(r.Role) {
			return nil, NewBizError(ErrCodeInvalidParam, "invalid role", nil)
		}
		if r.Role == model.AttendeeRoleOrganizer {
			return nil, NewBizError(ErrCodeInvalidParam, "cannot add organizer via API", nil)
		}
		// 去重
		if seen[r.UserID] {
			continue
		}
		seen[r.UserID] = true
		validReqs = append(validReqs, &r)
	}

	if len(validReqs) == 0 {
		return []*AttendeeResponse{}, nil
	}

	// 3. 数量上限校验
	count, err := s.attendeeRepo.CountByEvent(ctx, eventID, "")
	if err != nil {
		return nil, wrapAttendeeErr(err)
	}
	if count+int64(len(validReqs)) > maxAttendeesPerEvent {
		return nil, NewBizError(ErrCodeAttendeeCountExceed, "total attendee count exceeds limit", nil)
	}

	// 4. 构造实体
	attendees := make([]*model.ScheduleEventAttendee, 0, len(validReqs))
	for _, r := range validReqs {
		attendees = append(attendees, &model.ScheduleEventAttendee{
			EventID:        eventID,
			UserID:         r.UserID,
			Role:           r.Role,
			ResponseStatus: model.AttendeeResponsePending,
		})
	}

	// 5. 事务内批量写入
	err = s.txMgr.Transaction(ctx, func(txCtx context.Context) error {
		// 检查每个用户是否已是参与人
		for _, a := range attendees {
			existing, err := s.attendeeRepo.GetByEventAndUser(txCtx, a.EventID, a.UserID)
			if err != nil && !errors.Is(err, repo.ErrNotFound) {
				return wrapAttendeeErr(err)
			}
			if existing != nil {
				return NewBizError(ErrCodeAttendeeDuplicate, "user is already an attendee", nil)
			}
		}
		// 批量写入
		if err := s.attendeeRepo.BatchCreate(txCtx, attendees); err != nil {
			return wrapAttendeeErr(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 6. 转换响应
	result := make([]*AttendeeResponse, 0, len(attendees))
	for _, a := range attendees {
		result = append(result, attendeeToResponse(a))
	}
	return result, nil
}

// Update 更新参与人角色
//   仅 owner 可调用；组织者角色变更需特殊处理
func (s *scheduleAttendeeService) Update(ctx context.Context, actorID, attendeeID uint, role string) (*AttendeeResponse, error) {
	// 1. 参数校验
	if !validateAttendeeRole(role) {
		return nil, NewBizError(ErrCodeInvalidParam, "invalid role", nil)
	}

	// 2. 查询参与人记录
	attendee, err := s.attendeeRepo.GetByID(ctx, attendeeID)
	if err != nil {
		return nil, wrapAttendeeErr(err)
	}

	// 3. 权限校验
	event, err := s.eventRepo.GetByID(ctx, attendee.EventID)
	if err != nil {
		return nil, wrapEventErr(err)
	}
	if event.OwnerID != actorID {
		return nil, ErrForbidden
	}

	// 4. 组织者保护: 不可将组织者改为其他角色
	if attendee.Role == model.AttendeeRoleOrganizer && role != model.AttendeeRoleOrganizer {
		return nil, NewBizError(ErrCodeAttendeeOrganizerRequired, "organizer role cannot be changed", nil)
	}
	// 5. 不可将非组织者改为组织者（每事件唯一组织者）
	if attendee.Role != model.AttendeeRoleOrganizer && role == model.AttendeeRoleOrganizer {
		return nil, NewBizError(ErrCodeAttendeeOrganizerRequired, "cannot promote to organizer", nil)
	}

	// 6. 更新
	attendee.Role = role
	if err := s.attendeeRepo.Update(ctx, attendee); err != nil {
		return nil, wrapAttendeeErr(err)
	}

	return attendeeToResponse(attendee), nil
}

// Remove 移除参与人
//   校验: 操作者为 owner 或本人移除自己
//   组织者不可被移除
func (s *scheduleAttendeeService) Remove(ctx context.Context, actorID, attendeeID uint) error {
	// 1. 查询参与人
	attendee, err := s.attendeeRepo.GetByID(ctx, attendeeID)
	if err != nil {
		return wrapAttendeeErr(err)
	}

	// 2. 权限校验
	//    owner 可移除任何人，本人可移除自己
	event, err := s.eventRepo.GetByID(ctx, attendee.EventID)
	if err != nil {
		return wrapEventErr(err)
	}
	if event.OwnerID != actorID && attendee.UserID != actorID {
		return ErrForbidden
	}

	// 3. 组织者保护
	if attendee.Role == model.AttendeeRoleOrganizer {
		return NewBizError(ErrCodeAttendeeOrganizerRequired, "organizer cannot be removed", nil)
	}

	// 4. 删除
	if err := s.attendeeRepo.Delete(ctx, attendeeID); err != nil {
		return wrapAttendeeErr(err)
	}
	return nil
}

// GetByID 查询参与人记录详情
func (s *scheduleAttendeeService) GetByID(ctx context.Context, actorID, attendeeID uint) (*AttendeeResponse, error) {
	attendee, err := s.attendeeRepo.GetByID(ctx, attendeeID)
	if err != nil {
		return nil, wrapAttendeeErr(err)
	}

	// 权限校验: owner 或本人可查
	event, err := s.eventRepo.GetByID(ctx, attendee.EventID)
	if err != nil {
		return nil, wrapEventErr(err)
	}
	if event.OwnerID != actorID && attendee.UserID != actorID {
		return nil, ErrForbidden
	}

	return attendeeToResponse(attendee), nil
}

// ----------------------------------------------------------------------------
// 列表查询
// ----------------------------------------------------------------------------

// ListByEvent 查询某事件的参与人列表
func (s *scheduleAttendeeService) ListByEvent(ctx context.Context, actorID, eventID uint, req *PageRequest) (*AttendeeListResponse, error) {
	// 权限校验: owner 或参与人可查
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return nil, wrapEventErr(err)
	}
	if event.OwnerID != actorID {
		// 非owner须为参与人
		_, err := s.attendeeRepo.GetByEventAndUser(ctx, eventID, actorID)
		if err != nil {
			return nil, ErrForbidden
		}
	}

	params := toRepoQueryParams(*req)
	list, err := s.attendeeRepo.List(ctx, repo.AttendeeQueryOptions{
		EventID: eventID,
	}, params)
	if err != nil {
		return nil, wrapAttendeeErr(err)
	}
	return attendeesToListResponse(list.List, list.Total, list.Page, list.PageSize), nil
}

// ListByUser 查询用户参与的日程关联列表
func (s *scheduleAttendeeService) ListByUser(ctx context.Context, actorID uint, req *AttendeeListRequest) (*AttendeeListResponse, error) {
	// 仅可查自己的
	userID := actorID
	if req.UserID != 0 && req.UserID != actorID {
		return nil, ErrForbidden
	}
	params := toRepoQueryParams(req.PageRequest)
	list, err := s.attendeeRepo.List(ctx, repo.AttendeeQueryOptions{
		UserID:         userID,
		Role:           req.Role,
		ResponseStatus: req.ResponseStatus,
	}, params)
	if err != nil {
		return nil, wrapAttendeeErr(err)
	}
	return attendeesToListResponse(list.List, list.Total, list.Page, list.PageSize), nil
}

// ----------------------------------------------------------------------------
// 响应邀请
// ----------------------------------------------------------------------------

// Respond 响应事件邀请
//   actorID 必须为参与人本人
//   status: accepted / declined / tentative
func (s *scheduleAttendeeService) Respond(ctx context.Context, actorID, attendeeID uint, req *RespondInviteRequest) (*AttendeeResponse, error) {
	// 1. 参数校验
	if !validateResponseStatus(req.Status) {
		return nil, NewBizError(ErrCodeAttendeeResponseInvalid, "invalid response status", nil)
	}

	// 2. 查询参与人记录
	attendee, err := s.attendeeRepo.GetByID(ctx, attendeeID)
	if err != nil {
		return nil, wrapAttendeeErr(err)
	}

	// 3. 权限校验: 仅本人可响应
	if attendee.UserID != actorID {
		return nil, ErrForbidden
	}

	// 4. 更新响应状态
	if err := s.attendeeRepo.UpdateResponse(ctx, attendeeID, req.Status, time.Now()); err != nil {
		return nil, wrapAttendeeErr(err)
	}

	// 5. 同步内存对象返回
	attendee.ResponseStatus = req.Status
	now := time.Now()
	attendee.RespondedAt = &now

	return attendeeToResponse(attendee), nil
}

// ----------------------------------------------------------------------------
// 统计
// ----------------------------------------------------------------------------

// CountByEvent 统计事件参与人数（可按角色过滤）
func (s *scheduleAttendeeService) CountByEvent(ctx context.Context, actorID, eventID uint, role string) (int64, error) {
	// 权限校验: owner 或参与人可查
	event, err := s.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		return 0, wrapEventErr(err)
	}
	if event.OwnerID != actorID {
		_, err := s.attendeeRepo.GetByEventAndUser(ctx, eventID, actorID)
		if err != nil {
			return 0, ErrForbidden
		}
	}
	// 角色参数校验
	if role != "" && !validateAttendeeRole(role) {
		return 0, NewBizError(ErrCodeInvalidParam, "invalid role", nil)
	}
	return s.attendeeRepo.CountByEvent(ctx, eventID, role)
}
