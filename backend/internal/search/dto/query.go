package dto

import "github.com/Yogdunana/StarByte/backend/pkg/search"

// QueryRequest is POST /system/search/query.
type QueryRequest struct {
	Resource     string              `json:"resource" binding:"required"`
	Keyword      string              `json:"keyword"`
	Filters      *search.Group       `json:"filters"`
	Sorts        []search.Sort       `json:"sorts"`
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
	Cursor       string              `json:"cursor"`
	Aggregations []search.AggRequest `json:"aggregations"`
}

func (r QueryRequest) ToQuery() search.Query {
	return search.Query{
		Keyword: r.Keyword, Filters: r.Filters, Sorts: r.Sorts,
		Page: r.Page, PageSize: r.PageSize, Cursor: r.Cursor,
		Aggregations: r.Aggregations,
	}
}

// ResourceInfo is one searchable catalog entry.
type ResourceInfo struct {
	Code   string           `json:"code"`
	Name   string           `json:"name"`
	Fields []map[string]any `json:"fields"`
}
