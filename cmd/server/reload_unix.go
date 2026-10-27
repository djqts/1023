//go:build !windows

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func watchRuleReload(ctx context.Context, reload func() error, logger *slog.Logger) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP)
	defer signal.Stop(ch)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ch:
			if err := reload(); err != nil {
				logger.Warn("rule reload failed; keeping previous snapshot", "error", err)
			}
		}
	}
}
