package main

import "testing"

// TestSeedUsersUseDevPasswords 固定「哪些环境允许播种开发便利口令」的判定。
//
// 重点是 APP_ENV 未设置必须按生产处理：否则直接 `go run ./scripts`（绕过
// Makefile）且只注入 DB_* / JWT_SECRET 时，会把 admin/admin123（super_admin）
// 种进生产库 —— 与 loader 里「APP_ENV 空值按生产校验」的语义正相反。
func TestSeedUsersUseDevPasswords(t *testing.T) {
	tests := []struct {
		appEnv string
		want   bool
	}{
		{appEnv: "dev", want: true},
		{appEnv: "test", want: true},
		{appEnv: "", want: false},        // 未设置：按生产处理
		{appEnv: "prod", want: false},    // 生产
		{appEnv: "staging", want: false}, // 未知值：不放宽
		{appEnv: "DEV", want: false},     // 大小写敏感，不做模糊匹配
	}

	for _, tt := range tests {
		t.Run("APP_ENV="+tt.appEnv, func(t *testing.T) {
			t.Setenv("APP_ENV", tt.appEnv)
			if got := seedUsersUseDevPasswords(); got != tt.want {
				t.Fatalf("seedUsersUseDevPasswords() = %v，期望 %v", got, tt.want)
			}
		})
	}
}
