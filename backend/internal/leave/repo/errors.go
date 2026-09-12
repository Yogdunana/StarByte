package repo

import "errors"

// ErrNotPending 表示审批更新未命中 pending 行（并发下已被处理）。
var ErrNotPending = errors.New("leave application is not pending")
