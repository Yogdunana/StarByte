package service

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type JobHandler func(ctx context.Context, payload string, logf func(string)) error

type handlerMeta struct {
	fn   JobHandler
	desc string
	pub  bool
}

func builtinHandlers() map[string]handlerMeta {
	return map[string]handlerMeta{
		"noop": {fn: handleNoop, desc: "空操作，立即成功", pub: true},
		"echo": {fn: handleEcho, desc: "把 payload 写入执行日志", pub: true},
		"fail": {fn: handleFail, desc: "始终失败（用于重试/死信测试）", pub: true},
	}
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
	h, ok := builtinHandlers()[key]
	if !ok {
		return nil, false
	}
	return h.fn, true
}

func publicHandlers() []handlerInfo {
	src := builtinHandlers()
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
		return c.Err()
	}
}
