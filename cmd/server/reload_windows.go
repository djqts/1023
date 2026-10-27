//go:build windows

package main

import (
	"context"
	"log/slog"
)

// Windows 没有 SIGHUP；生产容器（Linux）使用 reload_unix.go。
func watchRuleReload(ctx context.Context, reload func() error, logger *slog.Logger) {
	_ = ctx
	_ = reload
	_ = logger
}
