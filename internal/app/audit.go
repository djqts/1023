package app

import (
	"context"
	"time"
)

// AuditSink 是批次审计的写出端口。
type AuditSink interface {
	// RecordBatch 记录一次批量识别审计。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//   - event: 批次摘要与可选明细。
	//
	// 输出:
	//   - 无；成功返回 nil。
	//
	// 错误:
	//   - 数据库/消息队列错误；调用方应降级并记录日志，不得阻断识别响应。
	RecordBatch(ctx context.Context, event AuditEvent) error
}

// AuditEvent 是一次批量识别的审计事件。
type AuditEvent struct {
	RequestID  string                  `json:"request_id" db:"-"`
	Batch      RecognitionBatchRecord  `json:"batch" db:"-"`
	Items      []RecognitionItemRecord `json:"items,omitempty" db:"-"`
	OccurredAt time.Time               `json:"occurred_at" db:"-"`
}
