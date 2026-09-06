package search

const (
	OpEq      = "eq"
	OpNe      = "ne"
	OpGt      = "gt"
	OpGte     = "gte"
	OpLt      = "lt"
	OpLte     = "lte"
	OpLike    = "like"
	OpPrefix  = "prefix"
	OpIn      = "in"
	OpBetween = "between"
	OpIsNull  = "is_null"
	OpNotNull = "not_null"

	LogicAnd = "and"
	LogicOr  = "or"

	FnCount = "count"
	FnSum   = "sum"
	FnAvg   = "avg"
	FnMin   = "min"
	FnMax   = "max"

	KindString = "string"
	KindNumber = "number"
	KindTime   = "time"
	KindBool   = "bool"

	MaxPageSize   = 100
	MaxOffset     = 10000
	MaxINValues   = 100
	MaxFilters    = 40
	MaxGroupDepth = 5
	DefaultPage   = 1
	DefaultSize   = 20
)

// Field is one whitelist column on a searchable resource.
type Field struct {
	Name       string `json:"name"`
	Column     string `json:"-"`
	Kind       string `json:"type"`
	Label      string `json:"label"`
	Searchable bool   `json:"searchable"`
	Filterable bool   `json:"filterable"`
	Sortable   bool   `json:"sortable"`
	Agg        bool   `json:"agg"`
}

// Schema describes a table that the engine may query.
type Schema struct {
	Code         string
	Name         string
	Table        string
	IDColumn     string
	FTSExpr      string
	Headline     string
	ExtraWhere   string
	ExtraArgs    []any
	RBACResource string // user / task / member / audit
	ScopeColumn  string // department_id when the table has one
	SelfSQL      string // parameterized predicate replacing middleware self (1 = 0)
	Fields       []Field
}

// Condition is one predicate. Value is JSON-decoded (number/string/bool/array).
type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

// Group is a nested AND/OR combination of conditions and child groups.
type Group struct {
	Logic      string      `json:"logic"`
	Conditions []Condition `json:"conditions"`
	Groups     []Group     `json:"groups"`
}

// Sort is one ORDER BY clause.
type Sort struct {
	Field string `json:"field"`
	Desc  bool   `json:"desc"`
}

// AggRequest is one aggregation. Cross-tab uses Row × Col.
type AggRequest struct {
	Name     string `json:"name"`
	Field    string `json:"field"`
	Fn       string `json:"fn"`
	Interval string `json:"interval"`
	Row      string `json:"row"`
	Col      string `json:"col"`
}

// Query is the unified search payload.
type Query struct {
	Keyword      string       `json:"keyword"`
	Filters      *Group       `json:"filters"`
	Sorts        []Sort       `json:"sorts"`
	Page         int          `json:"page"`
	PageSize     int          `json:"page_size"`
	Cursor       string       `json:"cursor"`
	Aggregations []AggRequest `json:"aggregations"`
}

// Result is one engine execution.
type Result struct {
	List         []map[string]any    `json:"list"`
	Total        int64               `json:"total"`
	Page         int                 `json:"page"`
	PageSize     int                 `json:"page_size"`
	NextCursor   string              `json:"next_cursor,omitempty"`
	HasMore      bool                `json:"has_more"`
	Aggregations map[string][]AggRow `json:"aggregations,omitempty"`
	ElapsedMs    int64               `json:"elapsed_ms"`
}

// AggRow is one bucket or cross-tab cell.
type AggRow struct {
	Key   any     `json:"key,omitempty"`
	Row   any     `json:"row,omitempty"`
	Col   any     `json:"col,omitempty"`
	Value float64 `json:"value"`
}

// Statement is compiled SQL ready for GORM Raw.
type Statement struct {
	SQL       string
	Args      []any
	CountSQL  string
	CountArgs []any
	Aggs      []AggStatement
	Limit     int
	Page      int
	PageSize  int
	Sorts     []Sort
}

// AggStatement is one aggregation query sharing the same WHERE as the list.
type AggStatement struct {
	Name string
	SQL  string
	Args []any
	Kind string // time, group, cross, scalar
}
