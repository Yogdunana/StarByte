package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSwaggerEnabled 钉死 fail-closed 口径：只有显式 APP_ENV=dev|test 才开启，
// 其余（含未设置的空值）一律关闭。
//
// 回归背景：原实现为 `os.Getenv("APP_ENV") != "prod"`，空值返回 true ——
// 而 loader 在 APP_ENV 为空时会补成 prod（按生产严格校验）。两者相反，
// 于是「按文档补齐密钥但没导出 APP_ENV」的部署会把 swagger 暴露出去。
// 本用例表中 APP_ENV="" 一项即为该回归的守卫，不要改回 want=true。
func TestSwaggerEnabled(t *testing.T) {
	cases := []struct {
		name   string
		appEnv string
		want   bool
	}{
		{name: "显式 dev", appEnv: "dev", want: true},
		{name: "显式 test", appEnv: "test", want: true},
		{name: "未设置（空值按生产）", appEnv: "", want: false},
		{name: "prod", appEnv: "prod", want: false},
		// 不在 loader 的 validEnvs 内，但也不该意外开启
		{name: "大小写不同（DEV）", appEnv: "DEV", want: false},
		{name: "staging", appEnv: "staging", want: false},
		{name: "大小写不同（Prod）", appEnv: "Prod", want: false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ENV", tt.appEnv)
			require.Equal(t, tt.want, swaggerEnabled())
		})
	}
}
