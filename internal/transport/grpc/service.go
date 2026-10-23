package grpcapi

import (
	"context"

	"banner-fingerprint/pkg/api"
)

// FingerprintService 是 gRPC 接入端口（P1，可选）。
type FingerprintService interface {
	// Fingerprint 处理批量识别 RPC。
	//
	// 输入:
	//   - ctx: RPC 上下文，含 deadline/metadata。
	//   - req: 批量请求。
	//
	// 输出:
	//   - api.BatchResponse: 七字段结果。
	//
	// 错误:
	//   - 由 gRPC 适配器映射为 InvalidArgument/Internal/Unavailable 等状态。
	Fingerprint(ctx context.Context, req api.BatchRequest) (api.BatchResponse, error)
}
