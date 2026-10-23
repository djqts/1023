package app

import (
	"context"
	"time"
)

// RuleAdminUseCase 是规则集管理应用服务端口（P1）。
type RuleAdminUseCase interface {
	// Publish 发布一个新的规则集版本，但不自动激活。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//   - cmd: 版本、内容、来源、操作人。
	//
	// 输出:
	//   - RuleSetVersion: 新版本元数据。
	//
	// 错误:
	//   - 内容非法、版本冲突、数据库错误、上下文取消。
	Publish(ctx context.Context, cmd PublishRuleSetCommand) (RuleSetVersion, error)

	// Activate 激活指定版本。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//   - version: 目标版本号。
	//
	// 输出:
	//   - RuleSetVersion: 激活后的版本元数据。
	//
	// 错误:
	//   - 版本不存在、规则校验失败、数据库错误、上下文取消。
	Activate(ctx context.Context, version string) (RuleSetVersion, error)

	// List 列出版本历史。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//
	// 输出:
	//   - []RuleSetVersion: 按时间倒序的版本列表。
	//
	// 错误:
	//   - 数据库错误、上下文取消。
	List(ctx context.Context) ([]RuleSetVersion, error)
}

// PublishRuleSetCommand 是发布规则集的命令。
type PublishRuleSetCommand struct {
	Version string `json:"version" db:"-"`
	Content []byte `json:"content" db:"-"`
	Source  string `json:"source" db:"-"`
	Actor   string `json:"actor" db:"-"`
}

// RuleSetVersion 是规则集版本元数据。
type RuleSetVersion struct {
	ID          uint64     `json:"id" db:"-"`
	Version     string     `json:"version" db:"-"`
	Digest      string     `json:"digest" db:"-"`
	Status      string     `json:"status" db:"-"`
	CreatedAt   time.Time  `json:"created_at" db:"-"`
	ActivatedAt *time.Time `json:"activated_at,omitempty" db:"-"`
}
