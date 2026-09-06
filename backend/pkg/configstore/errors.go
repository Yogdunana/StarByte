package configstore

import "errors"

// ErrNotFound 配置键不存在。
var ErrNotFound = errors.New("configstore: key not found")
