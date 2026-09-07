package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type JobHandler func(ctx context.Context, payload string, logf func(string)) error

type handlerMeta struct {
	fn   JobHandler
	desc string
	pub  bool
}

var (
	extraMu       sync.RWMutex
	extraHandlers = map[string]handlerMeta{}
)

func builtinHandlers() map[string]handlerMeta {
	return map[string]handlerMeta{
		"noop": {fn: handleNoop, desc: "空操作，立即成功", pub: true},
		"echo": {fn: handleEcho, desc: "把 payload 写入执行日志", pub: true},
		"fail": {fn: handleFail, desc: "始终失败（用于重试/死信测试）", pub: true},
	}
}

// RegisterHandler 供业务模块在启动时挂接处理器（如合同到期扫描）。
func RegisterHandler(key, desc string, fn JobHandler) {
	key = strings.TrimSpace(key)
	if key == "" || fn == nil {
		return
	}
	extraMu.Lock()
	defer extraMu.Unlock()
	extraHandlers[key] = handlerMeta{fn: fn, desc: desc, pub: true}
}

func handleNoop(_ context.Context, _ string, logf func(string)) error {
	logf("noop")
	return nil
}

func handleEcho(_ context.Context, payload string, logf func(string)) error {
	logf("echo: " + payload)
	return nil
}

func handleFail(_ context.Context, payload string, _ func(string)) error {
	msg := strings.TrimSpace(payload)
	if msg == "" {
		msg = "handler fail"
	}
	return fmt.Errorf("%s", msg)
}

func lookupHandler(key string) (JobHandler, bool) {
	if h, ok := builtinHandlers()[key]; ok {
		return h.fn, true
	}
	extraMu.RLock()
	defer extraMu.RUnlock()
	h, ok := extraHandlers[key]
	if !ok {
		return nil, false
	}
	return h.fn, true
}

func publicHandlers() []handlerInfo {
	src := builtinHandlers()
	extraMu.RLock()
	for k, v := range extraHandlers {
		src[k] = v
	}
	extraMu.RUnlock()
	out := make([]handlerInfo, 0, len(src))
	for k, v := range src {
		if v.pub {
			out = append(out, handlerInfo{Key: k, Description: v.desc})
		}
	}
	return out
}

type handlerInfo struct {
	Key         string
	Description string
}

func runWithTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	errCh := make(chan error, 1)
	go func() { errCh <- fn(c) }()
	select {
	case err := <-errCh:
		return err
	case <-c.Done():
		err := <-errCh
		if err != nil {
			return err
		}
		return c.Err()
	}
}
