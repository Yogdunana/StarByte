package model

type ReportType string

const (
	ReportTypeDaily   ReportType = "daily"
	ReportTypeWeekly  ReportType = "weekly"
	ReportTypeMonthly ReportType = "monthly"
)

func ValidReportType(reportType ReportType) bool {
	switch reportType {
	case ReportTypeDaily, ReportTypeWeekly, ReportTypeMonthly:
		return true
	default:
		return false
	}
}

type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"
	ReviewStatusApproved ReviewStatus = "approved"
	ReviewStatusRejected ReviewStatus = "rejected"
)

func ValidReviewStatus(status ReviewStatus) bool {
	switch status {
	case ReviewStatusPending, ReviewStatusApproved, ReviewStatusRejected:
		return true
	default:
		return false
	}
}
