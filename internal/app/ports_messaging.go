package app

import (
	"context"
	"time"

	"banner-fingerprint/internal/domain/fingerprint"
)

// JobMessage 是异步识别任务的队列消息。
type JobMessage struct {
	JobID          string              `json:"job_id" db:"-"`
	RequestID      string              `json:"request_id" db:"-"`
	IdempotencyKey string              `json:"idempotency_key" db:"-"`
	Items          []fingerprint.Input `json:"items" db:"-"`
	RuleDigest     string              `json:"rule_digest" db:"-"`
	Attempt        int                 `json:"attempt" db:"-"`
	CreatedAt      time.Time           `json:"created_at" db:"-"`
}

// AuditMessage 是审计事件的队列消息。
type AuditMessage struct {
	BatchID      string    `json:"batch_id" db:"-"`
	RequestID    string    `json:"request_id" db:"-"`
	ItemCount    int       `json:"item_count" db:"-"`
	UnknownCount int       `json:"unknown_count" db:"-"`
	RuleDigest   string    `json:"rule_digest" db:"-"`
	CreatedAt    time.Time `json:"created_at" db:"-"`
}

// RuleChangedMessage 是规则变更广播消息。
type RuleChangedMessage struct {
	Version   string    `json:"version" db:"-"`
	Digest    string    `json:"digest" db:"-"`
	Source    string    `json:"source" db:"-"`
	ChangedAt time.Time `json:"changed_at" db:"-"`
}

// JobQueue 是异步任务发布端口（P1）。
type JobQueue interface {
	// Publish 发布一条任务消息。
	//
	// 输入: ctx、msg。输出: 无。错误: MQ 不可用/发布未确认/上下文取消。
	Publish(ctx context.Context, msg JobMessage) error
}

// JobConsumer 是异步任务消费端口（P1）。
type JobConsumer interface {
	// Consume 消费任务消息，阻塞直到 ctx 取消。
	//
	// 输入:
	//   - ctx: 控制停止。
	//   - handler: 单条消息处理函数；返回 error 时由实现 nack/重试/DLQ。
	//
	// 输出:
	//   - 无；ctx 取消或消费循环退出时返回。
	//
	// 错误:
	//   - MQ 不可用、消费通道关闭、上下文取消。
	Consume(ctx context.Context, handler func(ctx context.Context, msg JobMessage) error) error
}
