package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/discipline/dto"
	"github.com/Yogdunana/StarByte/backend/internal/discipline/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisciplineApproveRevokeAppeal(t *testing.T) {
	mem := newMem()
	target := uuid.New()
	op := uuid.New()
	mem.users[target] = &model.NamedUser{ID: target, RealName: "张三"}
	mem.users[op] = &model.NamedUser{ID: op, RealName: "部长"}
	n := &captureNotify{}
	svc := New(mem, n, nil)
	ctx := context.Background()

	created, err := svc.Create(ctx, op, &dto.CreateRecordRequest{
		UserID: target.String(), Title: "迟到", Level: 1,
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, model.StatusPending, created.Status)
	assert.Equal(t, 0, n.n)

	id := uuid.MustParse(created.ID)
	approved, err := svc.Approve(ctx, op, id, "ok", nil)
	require.NoError(t, err)
	assert.Equal(t, model.StatusActive, approved.Status)
	assert.Equal(t, []string{"discipline_notice"}, n.codes)

	_, err = svc.Appeal(ctx, op, id, "不是我", nil)
	require.Error(t, err)
	assert.Equal(t, response.CodeDisciplineNoAccess, err.(*response.AppError).Code)

	appealed, err := svc.Appeal(ctx, target, id, "有误会", nil)
	require.NoError(t, err)
	assert.Equal(t, model.StatusAppealing, appealed.Status)
	assert.Len(t, appealed.Appeals, 1)

	_, err = svc.Appeal(ctx, target, id, "再申诉", nil)
	require.Error(t, err)
	assert.Equal(t, response.CodeDisciplineDupAppeal, err.(*response.AppError).Code)

	kept, err := svc.Approve(ctx, op, id, "维持", nil)
	require.NoError(t, err)
	assert.Equal(t, model.StatusActive, kept.Status)
	require.Len(t, kept.Appeals, 1)
	assert.Equal(t, model.AppealRejected, kept.Appeals[0].Status)

	again, err := svc.Appeal(ctx, target, id, "再次申诉", nil)
	require.NoError(t, err)
	assert.Equal(t, model.StatusAppealing, again.Status)

	revoked, err := svc.Revoke(ctx, op, id, "证据不足", nil)
	require.NoError(t, err)
	assert.Equal(t, model.StatusRevoked, revoked.Status)
	require.Len(t, revoked.Appeals, 2)
	gotStatus := map[int16]int{}
	for _, a := range revoked.Appeals {
		gotStatus[a.Status]++
	}
	assert.Equal(t, 1, gotStatus[model.AppealAccepted])
	assert.Equal(t, 1, gotStatus[model.AppealRejected])
	assert.Equal(t, []string{"discipline_notice", "discipline_notice", "discipline_revoked"}, n.codes)
}
