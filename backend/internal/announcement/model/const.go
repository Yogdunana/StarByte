package model

const (
	StatusDraft     int16 = 0
	StatusPublished int16 = 1
	StatusArchived  int16 = 2
)

const (
	CategoryAssociation = "association"
	CategoryActivity    = "activity"
	CategorySystem      = "system"
	CategoryPersonnel   = "personnel"
)

const (
	ContentMarkdown = "markdown"
	ContentHTML     = "html"
)

const MaxContentLen = 50000

func ValidCategory(v string) bool {
	switch v {
	case CategoryAssociation, CategoryActivity, CategorySystem, CategoryPersonnel:
		return true
	default:
		return false
	}
}

func ValidContentType(v string) bool {
	return v == ContentMarkdown || v == ContentHTML
}

func ValidStatus(v int16) bool {
	return v == StatusDraft || v == StatusPublished || v == StatusArchived
}
