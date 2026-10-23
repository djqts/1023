package app

import (
	"context"
	"time"

	"banner-fingerprint/internal/domain/fingerprint"
)

// FingerprintUseCase 是批量识别应用服务端口。
type FingerprintUseCase interface {
	// Execute 执行一批识别。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//   - cmd: 批量命令，Items 顺序即输出顺序。
	//
	// 输出:
	//   - BatchResult: 七字段结果、unknown 计数、耗时、规则摘要。
	//
	// 错误:
	//   - 请求超限、规则不可用、上下文取消等内部/边界错误。
	//   - 单条 unknown 不返回 error。
	Execute(ctx context.Context, cmd BatchCommand) (BatchResult, error)
}

// BatchCommand 是应用层批量识别命令。
type BatchCommand struct {
	RequestID string              `json:"request_id" db:"-"`
	Items     []fingerprint.Input `json:"items" db:"-"`
	Limits    Limits              `json:"limits" db:"-"`
}

// BatchResult 是应用层批量识别结果。
type BatchResult struct {
	RequestID  string               `json:"request_id" db:"-"`
	Results    []fingerprint.Result `json:"results" db:"-"`
	Total      int                  `json:"total" db:"-"`
	Unknown    int                  `json:"unknown" db:"-"`
	ElapsedMS  int64                `json:"elapsed_ms" db:"-"`
	RuleDigest string               `json:"rule_digest" db:"-"`
}

// Limits 是单批识别的边界。
type Limits struct {
	MaxItems       int           `json:"max_items" db:"-"`
	MaxBannerBytes int           `json:"max_banner_bytes" db:"-"`
	MaxBodyBytes   int64         `json:"max_body_bytes" db:"-"`
	RequestTimeout time.Duration `json:"request_timeout" db:"-"`
}
