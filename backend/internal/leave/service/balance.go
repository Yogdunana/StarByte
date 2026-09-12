package service

import (
	"context"

	"github.com/Yogdunana/StarByte/backend/internal/leave/dto"
	"github.com/Yogdunana/StarByte/backend/internal/leave/repo"
	"github.com/google/uuid"
)

func (s *leaveService) Balances(ctx context.Context, viewer Viewer, userID string, year int) ([]*dto.LeaveBalanceResponse, error) {
	target := viewer.UserID
	if userID != "" {
		parsed, err := uuid.Parse(userID)
		if err != nil {
			return nil, invalidTime("用户ID格式错误")
		}
		if parsed != viewer.UserID && !viewer.CanRead {
			return nil, noAccess("无权查看他人假期余额")
		}
		target = parsed
	}
	if year == 0 {
		year = bizYear(s.clock())
	}
	if err := s.ensureYearBalances(ctx, target, year); err != nil {
		return nil, err
	}
	rows, err := s.rows.GetLeaveBalancesByUser(ctx, target, year)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.LeaveBalanceResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapBalance(row))
	}
	return out, nil
}

func (s *leaveService) ensureYearBalances(ctx context.Context, userID uuid.UUID, year int) error {
	types, err := s.rows.GetAllLeaveTypes(ctx)
	if err != nil {
		return err
	}
	existing, err := s.rows.GetLeaveBalancesByUser(ctx, userID, year)
	if err != nil {
		return err
	}
	have := map[uuid.UUID]struct{}{}
	for _, row := range existing {
		have[row.LeaveTypeID] = struct{}{}
	}
	return s.rows.WithTx(ctx, func(tx repo.Repository) error {
		for i := range types {
			if _, ok := have[types[i].ID]; ok {
				continue
			}
			if err := s.createBalance(ctx, tx, userID, year, &types[i]); err != nil {
				return err
			}
		}
		return nil
	})
}
