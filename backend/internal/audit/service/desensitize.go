package service

import (
	"fmt"
	"regexp"
	"sync"
)

// sensitiveFields 需要脱敏的敏感字段名列表
// 此列表在 middleware 和 service 层之间共享，确保一致性
var sensitiveFields = []string{
	"password",
	"old_password",
	"new_password",
	"secret",
	"token",
	"access_token",
	"refresh_token",
}

// compiledPatterns 预编译所有敏感字段匹配正则，避免运行时竞态
var compiledPatterns struct {
	once     sync.Once
	patterns map[string]*regexp.Regexp
}

// initPatterns 初始化预编译正则（使用 sync.Once 保证线程安全）
func initPatterns() {
	compiledPatterns.patterns = make(map[string]*regexp.Regexp, len(sensitiveFields))
	for _, field := range sensitiveFields {
		pattern := fmt.Sprintf(`(?i)"%s"\s*:\s*"[^"]*"`, field)
		compiledPatterns.patterns[field] = regexp.MustCompile(pattern)
	}
}

// getPattern 返回指定敏感字段对应的预编译正则表达式
func getPattern(field string) *regexp.Regexp {
	compiledPatterns.once.Do(initPatterns)
	return compiledPatterns.patterns[field]
}
