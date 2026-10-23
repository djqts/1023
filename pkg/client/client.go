// Package client 定义调用 Banner 指纹识别服务的客户端契约。
package client

import (
	"context"
	"time"

	"banner-fingerprint/pkg/api"
)

// Client 是 server 的客户端端口。
type Client interface {
	// Fingerprint 发送一批扫描数据并返回逐条识别结果。
	//
	// 输入:
	//   - ctx: 控制请求取消与超时。
	//   - req: 批量请求，Items 顺序即响应顺序。
	//
	// 输出:
	//   - api.BatchResponse: 七字段结果；unknown 是正常结果。
	//
	// 错误:
	//   - 网络不可达、非 2xx、响应 JSON 解析失败。
	//   - context.Canceled / context.DeadlineExceeded。
	Fingerprint(ctx context.Context, req api.BatchRequest) (api.BatchResponse, error)

	// Health 查询 server 健康状态。
	//
	// 输入:
	//   - ctx: 控制请求取消与超时。
	//
	// 输出:
	//   - api.HealthResponse: 健康快照；503 时也可能返回该结构。
	//
	// 错误:
	//   - 网络不可达、非 2xx（含 503）、响应 JSON 解析失败。
	//   - context.Canceled / context.DeadlineExceeded。
	Health(ctx context.Context) (api.HealthResponse, error)
}

// Config 是客户端的连接与重试配置。
type Config struct {
	BaseURL      string        `json:"base_url" db:"-"`
	Timeout      time.Duration `json:"timeout" db:"-"`
	BatchSize    int           `json:"batch_size" db:"-"`
	MaxRetries   int           `json:"max_retries" db:"-"`
	RetryBackoff time.Duration `json:"retry_backoff" db:"-"`
	UserAgent    string        `json:"user_agent" db:"-"`
}
