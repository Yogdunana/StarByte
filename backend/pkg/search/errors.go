package search

import "errors"

var (
	ErrInvalidIdent   = errors.New("invalid identifier")
	ErrUnknownField   = errors.New("unknown field")
	ErrInvalidOp      = errors.New("invalid operator")
	ErrInvalidQuery   = errors.New("invalid query")
	ErrInvalidCursor  = errors.New("invalid cursor")
	ErrInvalidAgg     = errors.New("invalid aggregation")
	ErrDeepPagination = errors.New("offset too deep, use cursor")
	ErrUnknownTable   = errors.New("unknown resource")
)
