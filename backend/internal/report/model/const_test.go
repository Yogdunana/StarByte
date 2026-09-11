package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidReportType(t *testing.T) {
	for _, reportType := range []ReportType{ReportTypeDaily, ReportTypeWeekly, ReportTypeMonthly} {
		assert.True(t, ValidReportType(reportType))
	}
	assert.False(t, ValidReportType("custom"))
	assert.False(t, ValidReportType(""))
}

func TestValidReviewStatus(t *testing.T) {
	for _, status := range []ReviewStatus{ReviewStatusPending, ReviewStatusApproved, ReviewStatusRejected} {
		assert.True(t, ValidReviewStatus(status))
	}
	assert.False(t, ValidReviewStatus("draft"))
	assert.False(t, ValidReviewStatus(""))
}

func TestReportTableName(t *testing.T) {
	assert.Equal(t, "work_reports", (Report{}).TableName())
}
