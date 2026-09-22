package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/Yogdunana/StarByte/backend/pkg/utils"
)

// swapStdin 把 os.Stdin 换成内容为 content 的临时文件，测试结束自动还原。
// 不换的话 readLineFromStdin 在终端/管道下的行为取决于 go test 的调用方式，
// 测试会变得不可复现。
func swapStdin(t *testing.T, content string) {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stdin")
	if err != nil {
		t.Fatalf("create temp stdin: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp stdin: %v", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("seek temp stdin: %v", err)
	}
	old := os.Stdin
	os.Stdin = f
	t.Cleanup(func() {
		os.Stdin = old
		_ = f.Close()
	})
}

func TestValidateAdminUsername(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "admin", wantErr: false},
		{name: "sys_admin", wantErr: false},
		{name: "admin@smbu", wantErr: false},
		{name: "a", wantErr: false},
		{name: "", wantErr: true},                       // 空
		{name: "   ", wantErr: true},                    // 只有空白
		{name: strings.Repeat("a", 51), wantErr: true},  // 超 50
		{name: "admin 123", wantErr: true},              // 空格
		{name: "admin;drop", wantErr: true},             // 分号
		{name: "管理员", wantErr: true},                   // 非白名单字符
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAdminUsername(tt.name)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateAdminUsername(%q) err=%v, wantErr=%v", tt.name, err, tt.wantErr)
			}
		})
	}
}

// TestResolveNewPasswordOrder 固定「-value > 环境变量 > stdin > 自动生成」的优先级。
// 口令来源一多就容易搞反，而搞反的后果是运维以为设了新口令、实际是系统生成的。
func TestResolveNewPasswordOrder(t *testing.T) {
	const envPass = "FromEnv123"
	const stdinPass = "FromStdin123"
	const flagPass = "FromFlag123"

	t.Run("flag 优先于 env 和 stdin", func(t *testing.T) {
		t.Setenv(adminPasswordEnvVar, envPass)
		swapStdin(t, stdinPass+"\n")
		got, generated, err := resolveNewPassword(flagPass)
		if err != nil || generated || got != flagPass {
			t.Fatalf("got=%q generated=%v err=%v，期望原样用 flag 值", got, generated, err)
		}
	})

	t.Run("env 优先于 stdin", func(t *testing.T) {
		t.Setenv(adminPasswordEnvVar, envPass)
		swapStdin(t, stdinPass+"\n")
		got, generated, err := resolveNewPassword("")
		if err != nil || generated || got != envPass {
			t.Fatalf("got=%q generated=%v err=%v，期望用环境变量", got, generated, err)
		}
	})

	t.Run("stdin 优先于自动生成", func(t *testing.T) {
		t.Setenv(adminPasswordEnvVar, "")
		swapStdin(t, stdinPass+"\n")
		got, generated, err := resolveNewPassword("")
		if err != nil || generated || got != stdinPass {
			t.Fatalf("got=%q generated=%v err=%v，期望用 stdin", got, generated, err)
		}
	})

	t.Run("都没有则生成强随机口令", func(t *testing.T) {
		t.Setenv(adminPasswordEnvVar, "")
		swapStdin(t, "")
		got, generated, err := resolveNewPassword("")
		if err != nil || !generated {
			t.Fatalf("generated=%v err=%v，期望自动生成", generated, err)
		}
		if len(got) != adminGeneratedPasswordLen {
			t.Fatalf("生成的口令长度 %d，期望 %d", len(got), adminGeneratedPasswordLen)
		}
		// 生成的口令必须能通过登录侧的强度校验，否则改完还是登不进去。
		if !utils.ValidatePasswordStrength(got) {
			t.Fatalf("生成的口令未通过强度校验: %q", got)
		}
	})
}

func TestRunAdminModeRejectsUnknownMode(t *testing.T) {
	err := runAdminMode(nil, nil, "delete-everything", "admin", "")
	if err == nil {
		t.Fatal("未知 mode 应当报错")
	}
	if !strings.Contains(err.Error(), "未知的 -mode") {
		t.Fatalf("错误信息不含 mode 名: %v", err)
	}
}
