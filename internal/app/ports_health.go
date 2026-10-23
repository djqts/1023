package app

import (
	"context"
	"time"
)

// ProbeResult 是单个依赖的健康探测结果。
type ProbeResult struct {
	Name    string        `json:"name" db:"-"`
	Status  string        `json:"status" db:"-"`
	Latency time.Duration `json:"latency" db:"-"`
	Error   string        `json:"error,omitempty" db:"-"`
}

// RuleSetHealth 是健康检查中的规则集状态。
type RuleSetHealth struct {
	Count    int       `json:"count" db:"-"`
	Digest   string    `json:"digest" db:"-"`
	Version  string    `json:"version" db:"-"`
	LoadedAt time.Time `json:"loaded_at" db:"-"`
}

// HealthSnapshot 是服务健康快照。
type HealthSnapshot struct {
	Status       string        `json:"status" db:"-"`
	Version      string        `json:"version" db:"-"`
	Rules        RuleSetHealth `json:"rules" db:"-"`
	Dependencies []ProbeResult `json:"dependencies" db:"-"`
	Uptime       time.Duration `json:"uptime" db:"-"`
}

// HealthProbe 是依赖健康探测端口。
type HealthProbe interface {
	// Name 返回依赖名，用于健康 JSON 的 key。
	//
	// 输入: 无。输出: 依赖名。错误: 无。
	Name() string

	// Check 探测依赖健康。
	//
	// 输入: ctx。输出: 探测结果。错误: 不返回 error，失败写入 ProbeResult。
	Check(ctx context.Context) ProbeResult
}

// HealthService 聚合规则与依赖状态。
type HealthService interface {
	// Check 返回健康快照。
	//
	// 输入: ctx。输出: 快照。错误: 不返回 error；聚合失败体现为 degraded。
	Check(ctx context.Context) HealthSnapshot
}
