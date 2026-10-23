package httpapi

import (
	"context"

	"banner-fingerprint/pkg/api"
)

// FingerprintHandler 是 Gin 适配器依赖的 HTTP 应用端口。
type FingerprintHandler interface {
	// Batch 处理 POST /fingerprint。
	//
	// 输入:
	//   - ctx: 请求上下文。
	//   - req: 已反序列化的批量请求。
	//
	// 输出:
	//   - api.BatchResponse: 七字段结果，顺序与输入一致。
	//
	// 错误:
	//   - 由 Gin 适配器映射为 400/413/500/503。
	//   - unknown 不是错误。
	Batch(ctx context.Context, req api.BatchRequest) (api.BatchResponse, error)

	// Health 处理 GET /health。
	//
	// 输入:
	//   - ctx: 请求上下文。
	//
	// 输出:
	//   - api.HealthResponse: 健康快照；核心未就绪时由适配器映射 503。
	//
	// 错误:
	//   - 依赖探测失败时仍返回快照；仅在无法构造快照时返回 error。
	Health(ctx context.Context) (api.HealthResponse, error)
}

// RouteRegistrar 把 HTTP 端口注册到具体路由器（实现阶段由 Gin 适配）。
type RouteRegistrar interface {
	// Register 注册 /fingerprint 与 /health。
	//
	// 输入: handler。输出: 无。错误: 路由冲突/注册失败。
	Register(handler FingerprintHandler) error
}
