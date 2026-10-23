package app

import (
	"context"
	"time"
)

// RuleSnapshot 是 Redis 中缓存的规则快照。
type RuleSnapshot struct {
	Version   string    `json:"version" db:"-"`
	Digest    string    `json:"digest" db:"-"`
	Content   []byte    `json:"content" db:"-"`
	Source    string    `json:"source" db:"-"`
	UpdatedAt time.Time `json:"updated_at" db:"-"`
}

// CachedResult 是结果缓存值；不含 ip/port，命中后由请求回填。
type CachedResult struct {
	Protocol   string  `json:"protocol" db:"-"`
	Product    string  `json:"product" db:"-"`
	Version    string  `json:"version" db:"-"`
	OSHint     string  `json:"os_hint" db:"-"`
	Confidence float64 `json:"confidence" db:"-"`
}

// RateLimitDecision 是限流判定结果。
type RateLimitDecision struct {
	Allowed   bool      `json:"allowed" db:"-"`
	Remaining int       `json:"remaining" db:"-"`
	ResetAt   time.Time `json:"reset_at" db:"-"`
}

// JobStatus 是任务状态缓存值。
type JobStatus struct {
	JobID     string    `json:"job_id" db:"-"`
	Status    string    `json:"status" db:"-"`
	ErrorCode string    `json:"error_code,omitempty" db:"-"`
	UpdatedAt time.Time `json:"updated_at" db:"-"`
}

// RuleChangeEvent 是规则变更通知。
type RuleChangeEvent struct {
	Version   string    `json:"version" db:"-"`
	Digest    string    `json:"digest" db:"-"`
	Source    string    `json:"source" db:"-"`
	ChangedAt time.Time `json:"changed_at" db:"-"`
}

// RuleChangeHandler 处理规则变更通知。
type RuleChangeHandler interface {
	// Handle 处理一次规则变更。
	//
	// 输入: ctx、event。输出: 无。错误: 处理失败，由通知实现决定是否重试。
	Handle(ctx context.Context, event RuleChangeEvent) error
}

// RuleCache 是规则快照缓存端口。
type RuleCache interface {
	// GetActive 读取 active 规则快照。
	//
	// 输入: ctx。输出: 快照、是否命中、错误。错误: Redis 故障/上下文取消。
	GetActive(ctx context.Context) (RuleSnapshot, bool, error)

	// SetActive 写入 active 规则快照。
	//
	// 输入: ctx、snapshot、ttl。输出: 无。错误: Redis 故障/上下文取消。
	SetActive(ctx context.Context, snapshot RuleSnapshot, ttl time.Duration) error

	// DeleteActive 删除 active 快照。
	//
	// 输入: ctx。输出: 无。错误: Redis 故障/上下文取消。
	DeleteActive(ctx context.Context) error
}

// ResultCache 是识别结果缓存端口（P1，默认关闭）。
type ResultCache interface {
	// Get 读取缓存结果。
	//
	// 输入: ctx、key。输出: 结果、是否命中、错误。错误: Redis 故障/上下文取消。
	Get(ctx context.Context, key string) (CachedResult, bool, error)

	// Set 写入缓存结果。
	//
	// 输入: ctx、key、value、ttl。输出: 无。错误: Redis 故障/上下文取消。
	Set(ctx context.Context, key string, value CachedResult, ttl time.Duration) error
}

// RateLimiter 是限流端口（P1，公开暴露时启用）。
type RateLimiter interface {
	// Allow 判断一个 key 在窗口内是否允许通过。
	//
	// 输入: ctx、key、limit、window。输出: 判定结果。错误: Redis 故障/上下文取消。
	Allow(ctx context.Context, key string, limit int, window time.Duration) (RateLimitDecision, error)
}

// IdempotencyStore 是异步任务幂等快速判重端口（P1）。
type IdempotencyStore interface {
	// Acquire 尝试占用幂等键。
	//
	// 输入: ctx、key、ttl。输出: 是否占用成功。错误: Redis 故障/上下文取消。
	Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// Release 释放幂等键。
	//
	// 输入: ctx、key。输出: 无。错误: Redis 故障/上下文取消。
	Release(ctx context.Context, key string) error
}

// JobStatusCache 是任务状态查询缓存端口（P1）。
type JobStatusCache interface {
	// Get 读取任务状态缓存。
	//
	// 输入: ctx、jobID。输出: 状态、是否命中、错误。错误: Redis 故障/上下文取消。
	Get(ctx context.Context, jobID string) (JobStatus, bool, error)

	// Set 写入任务状态缓存。
	//
	// 输入: ctx、jobID、status、ttl。输出: 无。错误: Redis 故障/上下文取消。
	Set(ctx context.Context, jobID string, status JobStatus, ttl time.Duration) error
}

// RuleChangeNotifier 是规则变更发布/订阅端口（P1）。
type RuleChangeNotifier interface {
	// Publish 发布规则变更事件。
	//
	// 输入: ctx、event。输出: 无。错误: Redis 故障/上下文取消。
	Publish(ctx context.Context, event RuleChangeEvent) error

	// Subscribe 订阅规则变更事件，阻塞直到 ctx 取消。
	//
	// 输入: ctx、handler。输出: 无。错误: Redis 故障/上下文取消。
	Subscribe(ctx context.Context, handler RuleChangeHandler) error
}
