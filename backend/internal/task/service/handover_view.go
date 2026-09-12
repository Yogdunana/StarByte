package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/Yogdunana/StarByte/backend/internal/task/dto"
	"github.com/Yogdunana/StarByte/backend/internal/task/model"
	"github.com/Yogdunana/StarByte/backend/internal/workflow/engine"
)

func (s *taskService) handoverSnapshot(ctx context.Context, t *model.Task, request *model.TaskTransfer, actor uuid.UUID) (*dto.HandoverResponse, error) {
	if request == nil {
		return nil, nil
	}
	person := func(id uuid.UUID) (dto.Person, error) {
		if id == uuid.Nil {
			return dto.Person{}, nil
		}
		u, err := s.tasks.GetUser(ctx, id)
		return dto.Person{ID: id.String(), Name: displayName(u)}, err
	}
	from, err := person(request.FromUserID)
	if err != nil {
		return nil, err
	}
	to, err := person(request.ToUserID)
	if err != nil {
		return nil, err
	}
	out := &dto.HandoverResponse{
		ID: request.ID.String(), TaskID: t.ID.String(), TaskTitle: t.Title, Kind: request.Kind,
		Status: request.Status, Revision: request.Revision, From: from, To: to, Reason: request.Reason,
		Requirements: engine.TaskTransferRoles(request.Kind), CanSign: []string{}, Signatures: []dto.HandoverSign{},
		CreatedAt: request.CreatedAt,
	}
	if request.SourceDepartmentID != uuid.Nil && s.transfers != nil {
		if dept, err := s.transfers.Department(ctx, request.SourceDepartmentID); err != nil {
			return nil, err
		} else if dept != nil {
			out.SourceDepartment = dept.Name
		}
	}
	if request.TargetDepartmentID != uuid.Nil && s.transfers != nil {
		if dept, err := s.transfers.Department(ctx, request.TargetDepartmentID); err != nil {
			return nil, err
		} else if dept != nil {
			out.TargetDepartment = dept.Name
		}
	}
	if s.transfers != nil {
		rows, err := s.transfers.Signatures(ctx, request.ID)
		if err != nil {
			return nil, err
		}
		signed := map[string]bool{}
		for _, row := range rows {
			signer, err := person(row.SignerID)
			if err != nil {
				return nil, err
			}
			out.Signatures = append(out.Signatures, dto.HandoverSign{
				ID: row.ID.String(), Requirement: row.Requirement, Signer: signer,
				SignerRole: row.SignerRole, Waived: row.Waived, Decision: row.Decision,
				Comment: row.Comment, CreatedAt: row.CreatedAt,
			})
			if row.Decision == "approve" {
				signed[row.Requirement] = true
			}
		}
		if request.Status == "pending" {
			if signer, err := s.transfers.Actor(ctx, actor); err != nil {
				return nil, err
			} else {
				for _, role := range out.Requirements {
					if signed[role] {
						continue
					}
					if actual, _ := transferAuthority(signer, request, role); actual != "" {
						out.CanSign = append(out.CanSign, role)
					}
				}
			}
		}
	}
	return out, nil
}
