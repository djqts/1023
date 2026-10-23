// Package observability 定义日志与请求上下文端口。
package observability

import "context"

// Logger 是结构化日志端口；实现阶段用 slog 的 JSON Handler 适配。
type Logger interface {
	// Debug 记录调试日志。
	// 输入: ctx、msg、args。输出: 无。错误: 不返回 error。
	Debug(ctx context.Context, msg string, args ...any)

	// Info 记录信息日志。
	// 输入: ctx、msg、args。输出: 无。错误: 不返回 error。
	Info(ctx context.Context, msg string, args ...any)

	// Warn 记录告警日志。
	// 输入: ctx、msg、args。输出: 无。错误: 不返回 error。
	Warn(ctx context.Context, msg string, args ...any)

	// Error 记录错误日志。
	// 输入: ctx、msg、err、args。输出: 无。错误: 不返回 error。
	Error(ctx context.Context, msg string, err error, args ...any)

	// With 派生带固定字段的 logger。
	// 输入: args。输出: 新的 Logger。错误: 不返回 error。
	With(args ...any) Logger
}
