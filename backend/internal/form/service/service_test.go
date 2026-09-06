package service

import (
	"context"
	"testing"

	"github.com/Yogdunana/StarByte/backend/internal/form/dto"
	"github.com/Yogdunana/StarByte/backend/internal/form/model"
	"github.com/Yogdunana/StarByte/backend/pkg/response"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func sampleFields() []model.FormField {
	minLen := 2
	return []model.FormField{
		{Name: "real_name", Label: "姓名", Type: "text", Required: true, Validation: &model.FieldValidation{MinLength: &minLen}},
		{Name: "grade", Label: "年级", Type: "select", Required: true, Options: []model.FieldOption{{Label: "大一", Value: 1}, {Label: "大二", Value: 2}}},
		{Name: "intro", Label: "介绍", Type: "textarea", VisibleWhen: &model.VisibleWhen{Field: "grade", Operator: "==", Value: 2}},
	}
}

func TestFormCRUDAndSubmit(t *testing.T) {
	mem := newMemRepo()
	uid := uuid.New()
	mem.names[uid] = "张三管理员"
	svc := New(mem)
	ctx := context.Background()

	created, err := svc.Create(ctx, uid, dto.CreateFormRequest{Name: "入会申请表", Fields: sampleFields()})
	require.NoError(t, err)
	require.Equal(t, int16(0), created.Status)
	id := uuid.MustParse(created.ID)

	_, err = svc.Submit(ctx, uid, id, map[string]interface{}{"real_name": "张三", "grade": 2})
	require.Error(t, err)
	require.Equal(t, response.CodeFormNotPublished, err.(*response.AppError).Code)

	pub := model.StatusPublished
	_, err = svc.Update(ctx, uid, id, dto.UpdateFormRequest{Status: &pub})
	require.NoError(t, err)

	_, err = svc.Submit(ctx, uid, id, map[string]interface{}{"grade": 1})
	require.Error(t, err)
	require.Equal(t, response.CodeFormFieldRequired, err.(*response.AppError).Code)

	out, err := svc.Submit(ctx, uid, id, map[string]interface{}{"real_name": "张", "grade": 1})
	require.Error(t, err)
	require.Equal(t, response.CodeFormFieldInvalid, err.(*response.AppError).Code)
	require.Nil(t, out)

	ok, err := svc.Submit(ctx, uid, id, map[string]interface{}{"real_name": "张三", "grade": 2, "intro": "你好"})
	require.NoError(t, err)
	require.NotEmpty(t, ok.SubmissionID)

	list, total, _, _, err := svc.ListSubmissions(ctx, id, dto.SubmissionQuery{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	require.NotNil(t, list[0].SubmittedBy)
	require.Equal(t, "张三管理员", list[0].SubmittedBy.Name)

	items, _, _, _, err := svc.List(ctx, dto.ListQuery{Page: 1, PageSize: 10}, false)
	require.NoError(t, err)
	require.Len(t, items, 1)
}

func TestSchemaAndDuplicateName(t *testing.T) {
	svc := New(newMemRepo())
	ctx := context.Background()
	uid := uuid.New()
	_, err := svc.Create(ctx, uid, dto.CreateFormRequest{
		Name: "A", Fields: []model.FormField{{Name: "1bad", Label: "x", Type: "text"}},
	})
	require.Error(t, err)
	require.Equal(t, response.CodeFormInvalidSchema, err.(*response.AppError).Code)

	_, err = svc.Create(ctx, uid, dto.CreateFormRequest{Name: "报名", Fields: sampleFields()})
	require.NoError(t, err)
	_, err = svc.Create(ctx, uid, dto.CreateFormRequest{Name: "报名", Fields: sampleFields()})
	require.Error(t, err)
	require.Equal(t, response.CodeFormNameExists, err.(*response.AppError).Code)
}

func TestHiddenRequiredWhenDependencyMissing(t *testing.T) {
	svc := New(newMemRepo())
	ctx := context.Background()
	uid := uuid.New()
	fields := []model.FormField{
		{Name: "score", Label: "分数", Type: "number"},
		{
			Name: "reason", Label: "原因", Type: "text", Required: true,
			VisibleWhen: &model.VisibleWhen{Field: "score", Operator: ">", Value: float64(80)},
		},
	}
	created, err := svc.Create(ctx, uid, dto.CreateFormRequest{Name: "评分表", Fields: fields})
	require.NoError(t, err)
	id := uuid.MustParse(created.ID)
	pub := model.StatusPublished
	_, err = svc.Update(ctx, uid, id, dto.UpdateFormRequest{Status: &pub})
	require.NoError(t, err)

	_, err = svc.Submit(ctx, uid, id, map[string]interface{}{})
	require.NoError(t, err)

	_, err = svc.Submit(ctx, uid, id, map[string]interface{}{"score": 90})
	require.Error(t, err)
	require.Equal(t, response.CodeFormFieldRequired, err.(*response.AppError).Code)
}

func TestInvalidPatternRejectedOnSave(t *testing.T) {
	svc := New(newMemRepo())
	_, err := svc.Create(context.Background(), uuid.New(), dto.CreateFormRequest{
		Name: "坏正则",
		Fields: []model.FormField{
			{Name: "phone", Label: "手机", Type: "text", Validation: &model.FieldValidation{Pattern: "["}},
		},
	})
	require.Error(t, err)
	require.Equal(t, response.CodeFormInvalidSchema, err.(*response.AppError).Code)
}

func TestInvalidPatternSkippedOnSubmit(t *testing.T) {
	f := model.FormField{
		Name: "phone", Label: "手机", Type: "text",
		Validation: &model.FieldValidation{Pattern: "["},
	}
	require.NoError(t, validateValue(f, "13800000000"))
}

func TestCompareNumMissingLeftIsHidden(t *testing.T) {
	f := model.FormField{
		Name: "reason", Label: "原因", Type: "text", Required: true,
		VisibleWhen: &model.VisibleWhen{Field: "score", Operator: ">", Value: float64(80)},
	}
	require.False(t, isVisible(f, map[string]interface{}{}))
	require.False(t, isVisible(f, map[string]interface{}{"score": nil}))
	require.True(t, isVisible(f, map[string]interface{}{"score": 90}))
	require.False(t, isVisible(f, map[string]interface{}{"score": 10}))
}
