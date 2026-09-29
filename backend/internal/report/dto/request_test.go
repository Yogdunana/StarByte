package dto

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateReportRequestValidation(t *testing.T) {
	validate := newBindingValidator()
	now := time.Now()
	later := now.AddDate(0, 0, 1)

	valid := CreateReportRequest{
		ReportType:  "weekly",
		PeriodStart: &now,
		PeriodEnd:   &later,
	}
	require.NoError(t, validate.Struct(valid))

	invalidType := valid
	invalidType.ReportType = "custom"
	assert.Error(t, validate.Struct(invalidType))

	missingPeriod := valid
	missingPeriod.PeriodStart = nil
	assert.Error(t, validate.Struct(missingPeriod))

	missingPeriod = valid
	missingPeriod.PeriodEnd = nil
	assert.Error(t, validate.Struct(missingPeriod))

	reversedPeriod := valid
	reversedPeriod.PeriodStart = &later
	reversedPeriod.PeriodEnd = &now
	assert.Error(t, validate.Struct(reversedPeriod))
}

func TestUpdateReportRequestRequiresAtLeastOneField(t *testing.T) {
	validate := newBindingValidator()

	assert.Error(t, validate.Struct(UpdateReportRequest{}))
	empty := ""
	require.NoError(t, validate.Struct(UpdateReportRequest{WorkContent: &empty}))
}

func TestReviewReportRequestValidation(t *testing.T) {
	validate := newBindingValidator()

	for _, status := range []string{"approved", "rejected"} {
		require.NoError(t, validate.Struct(ReviewReportRequest{ReviewStatus: status}))
	}
	assert.Error(t, validate.Struct(ReviewReportRequest{ReviewStatus: "pending"}))
	assert.Error(t, validate.Struct(ReviewReportRequest{ReviewStatus: "draft"}))
}

func TestListReportRequestValidation(t *testing.T) {
	validate := newBindingValidator()
	userID := "9cf123fe-4755-4f72-b3c4-69a01608894f"

	require.NoError(t, validate.Struct(ListReportRequest{
		ReportType: "daily", ReviewStatus: "pending", UserID: userID, DepartmentID: userID,
	}))
	assert.Error(t, validate.Struct(ListReportRequest{ReportType: "custom"}))
	assert.Error(t, validate.Struct(ListReportRequest{ReviewStatus: "draft"}))
	assert.Error(t, validate.Struct(ListReportRequest{PageSize: 101}))
	assert.Error(t, validate.Struct(ListReportRequest{UserID: "not-a-uuid"}))
	assert.Error(t, validate.Struct(ListReportRequest{DepartmentID: "not-a-uuid"}))
}

func newBindingValidator() *validator.Validate {
	validate := validator.New()
	validate.SetTagName("binding")
	return validate
}
