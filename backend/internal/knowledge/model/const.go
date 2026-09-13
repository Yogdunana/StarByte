package model

import "regexp"

const (
	KindPage = "page"
	KindDoc  = "doc"
)

const (
	VisibilityRole          = "role"
	VisibilityPublic        = "public"
	VisibilityAuthenticated = "authenticated"
	VisibilityPermission    = "permission"
)

const (
	StatusDraft     int16 = 0
	StatusPublished int16 = 1
)

const MaxContentLen = 200000
const MaxTitleLen = 200
const MaxSummaryLen = 500
const MaxSlugLen = 80

var slugRE = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func ValidKind(v string) bool {
	return v == KindPage || v == KindDoc
}

func ValidVisibility(v string) bool {
	switch v {
	case VisibilityPublic, VisibilityAuthenticated, VisibilityPermission, VisibilityRole:
		return true
	default:
		return false
	}
}

func ValidStatus(v int16) bool {
	return v == StatusDraft || v == StatusPublished
}

func ValidSlug(v string) bool {
	if v == "" || len(v) > MaxSlugLen {
		return false
	}
	return slugRE.MatchString(v)
}

func PublicPath(kind, slug string) string {
	if kind == KindPage {
		return "/" + slug
	}
	return "/docs/" + slug
}
