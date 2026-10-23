package app

import (
	"context"
	"time"

	"banner-fingerprint/internal/domain/fingerprint"
)

// JobUseCase 是异步识别任务应用服务端口（P1）。
type JobUseCase interface {
	// Create 创建异步识别任务。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//   - cmd: 请求 ID、幂等键、待识别条目。
	//
	// 输出:
	//   - AsyncJob: 任务 ID 与初始状态。
	//
	// 错误:
	//   - 幂等键冲突、队列不可用、数据库错误、上下文取消。
	Create(ctx context.Context, cmd CreateJobCommand) (AsyncJob, error)

	// Get 查询任务状态。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//   - jobID: 任务 ID。
	//
	// 输出:
	//   - AsyncJob: 当前状态与错误码。
	//
	// 错误:
	//   - 任务不存在、数据库/缓存错误、上下文取消。
	Get(ctx context.Context, jobID string) (AsyncJob, error)
}

// CreateJobCommand 是创建异步任务的命令。
type CreateJobCommand struct {
	RequestID      string              `json:"request_id" db:"-"`
	IdempotencyKey string              `json:"idempotency_key" db:"-"`
	Items          []fingerprint.Input `json:"items" db:"-"`
}

// AsyncJob 是异步任务的应用层视图。
type AsyncJob struct {
	ID        string    `json:"id" db:"-"`
	Status    string    `json:"status" db:"-"`
	ErrorCode string    `json:"error_code,omitempty" db:"-"`
	CreatedAt time.Time `json:"created_at" db:"-"`
	UpdatedAt time.Time `json:"updated_at" db:"-"`
}
