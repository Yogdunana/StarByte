package service

import (
	"context"

	userrepo "github.com/Yogdunana/StarByte/backend/internal/user/repo"
	"github.com/google/uuid"
)

type userDepartments struct{ users userrepo.UserRepo }

func NewUserDepartments(users userrepo.UserRepo) UserLookup {
	return &userDepartments{users: users}
}

func (u *userDepartments) DepartmentOf(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	if u.users == nil {
		return nil, nil
	}
	row, err := u.users.GetByID(ctx, userID)
	if err != nil || row == nil {
		return nil, err
	}
	return row.DepartmentID, nil
}
