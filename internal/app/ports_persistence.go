package app

import (
	"context"
	"time"
)

// RuleSetRecord 是规则集持久化记录的应用层视图。
type RuleSetRecord struct {
	ID          uint64     `json:"id" db:"id"`
	Version     string     `json:"version" db:"version"`
	Digest      string     `json:"digest" db:"digest"`
	Content     string     `json:"content" db:"content"`
	Source      string     `json:"source" db:"source"`
	Status      string     `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ActivatedAt *time.Time `json:"activated_at,omitempty" db:"activated_at"`
}

// RecognitionBatchRecord 是批次审计摘要记录。
type RecognitionBatchRecord struct {
	ID           uint64    `json:"id" db:"id"`
	RequestID    string    `json:"request_id" db:"request_id"`
	ItemCount    int       `json:"item_count" db:"item_count"`
	UnknownCount int       `json:"unknown_count" db:"unknown_count"`
	LatencyMS    int64     `json:"latency_ms" db:"latency_ms"`
	RuleDigest   string    `json:"rule_digest" db:"rule_digest"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// RecognitionItemRecord 是单条识别明细审计记录。
type RecognitionItemRecord struct {
	ID         uint64    `json:"id" db:"id"`
	BatchID    uint64    `json:"batch_id" db:"batch_id"`
	IPHash     string    `json:"ip_hash" db:"ip_hash"`
	Port       int       `json:"port" db:"port"`
	Protocol   string    `json:"protocol" db:"protocol"`
	Product    string    `json:"product" db:"product"`
	Version    string    `json:"version" db:"version"`
	OSHint     string    `json:"os_hint" db:"os_hint"`
	Confidence float64   `json:"confidence" db:"confidence"`
	RuleID     string    `json:"rule_id" db:"rule_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// AsyncJobRecord 是异步任务持久化记录。
type AsyncJobRecord struct {
	ID             string    `json:"id" db:"id"`
	IdempotencyKey string    `json:"idempotency_key" db:"idempotency_key"`
	Status         string    `json:"status" db:"status"`
	InputRef       string    `json:"input_ref" db:"input_ref"`
	ResultRef      string    `json:"result_ref" db:"result_ref"`
	ErrorCode      string    `json:"error_code" db:"error_code"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// OutboxEventRecord 是事务性发件箱记录。
type OutboxEventRecord struct {
	ID            uint64     `json:"id" db:"id"`
	AggregateType string     `json:"aggregate_type" db:"aggregate_type"`
	AggregateID   string     `json:"aggregate_id" db:"aggregate_id"`
	EventType     string     `json:"event_type" db:"event_type"`
	Payload       []byte     `json:"payload" db:"payload"`
	Status        string     `json:"status" db:"status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	PublishedAt   *time.Time `json:"published_at,omitempty" db:"published_at"`
}

// RuleRepository 是规则集元数据持久化端口。
type RuleRepository interface {
	// GetActive 返回当前激活的规则集记录。
	//
	// 输入: ctx。输出: 记录。错误: 未找到/数据库错误/上下文取消。
	GetActive(ctx context.Context) (RuleSetRecord, error)

	// Save 保存一个规则集版本（不激活）。
	//
	// 输入: ctx、record。输出: 无。错误: 版本冲突/数据库错误/上下文取消。
	Save(ctx context.Context, record RuleSetRecord) error

	// Activate 激活指定 ID 的规则集。
	//
	// 输入: ctx、id。输出: 无。错误: 记录不存在/数据库错误/上下文取消。
	Activate(ctx context.Context, id uint64) error
}

// AuditRepository 是批次审计持久化端口。
type AuditRepository interface {
	// InsertBatch 写入批次摘要。
	//
	// 输入: ctx、record。输出: 无。错误: 数据库错误/上下文取消。
	InsertBatch(ctx context.Context, record RecognitionBatchRecord) error

	// InsertItems 批量写入识别明细。
	//
	// 输入: ctx、records。输出: 无。错误: 数据库错误/上下文取消。
	InsertItems(ctx context.Context, records []RecognitionItemRecord) error

	// GetBatch 查询批次摘要。
	//
	// 输入: ctx、id。输出: 记录。错误: 未找到/数据库错误/上下文取消。
	GetBatch(ctx context.Context, id uint64) (RecognitionBatchRecord, error)
}

// JobRepository 是异步任务持久化端口。
type JobRepository interface {
	// Create 创建任务记录。
	//
	// 输入: ctx、record。输出: 无。错误: 幂等键冲突/数据库错误/上下文取消。
	Create(ctx context.Context, record AsyncJobRecord) error

	// UpdateStatus 更新任务状态与错误码。
	//
	// 输入: ctx、id、status、errorCode。输出: 无。错误: 未找到/数据库错误/上下文取消。
	UpdateStatus(ctx context.Context, id, status, errorCode string) error

	// Get 查询任务记录。
	//
	// 输入: ctx、id。输出: 记录。错误: 未找到/数据库错误/上下文取消。
	Get(ctx context.Context, id string) (AsyncJobRecord, error)
}

// OutboxRepository 是事务性发件箱端口。
type OutboxRepository interface {
	// Append 在事务内追加一条待发布事件。
	//
	// 输入: ctx、record。输出: 无。错误: 数据库错误/上下文取消。
	Append(ctx context.Context, record OutboxEventRecord) error

	// FetchPending 拉取待发布事件。
	//
	// 输入: ctx、limit。输出: 事件列表。错误: 数据库错误/上下文取消。
	FetchPending(ctx context.Context, limit int) ([]OutboxEventRecord, error)

	// MarkPublished 标记事件已发布。
	//
	// 输入: ctx、id。输出: 无。错误: 未找到/数据库错误/上下文取消。
	MarkPublished(ctx context.Context, id uint64) error
}

// UnitOfWork 定义本地事务边界。
type UnitOfWork interface {
	// Transaction 在同一个事务中执行 fn。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//   - fn: 事务内回调，必须使用传入的 txCtx。
	//
	// 输出:
	//   - 无；fn 返回 nil 时提交，否则回滚。
	//
	// 错误:
	//   - fn 返回的错误、提交/回滚错误、上下文取消。
	Transaction(ctx context.Context, fn func(txCtx context.Context) error) error
}
