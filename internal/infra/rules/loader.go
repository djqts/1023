package rules

import (
	"context"
	"time"
)

// Source 标识规则来源。
type Source string

// 规则来源取值。
const (
	SourceYAML  Source = "yaml"
	SourceMySQL Source = "mysql"
	SourceRedis Source = "redis"
)

// RawRuleSet 是规则加载器返回的原始规则快照。
type RawRuleSet struct {
	Version  string    `json:"version" db:"-"`
	Digest   string    `json:"digest" db:"-"`
	Content  []byte    `json:"content" db:"-"`
	Source   Source    `json:"source" db:"-"`
	LoadedAt time.Time `json:"loaded_at" db:"-"`
}

// Loader 是规则源加载端口。
type Loader interface {
	// Load 读取一次规则快照。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//
	// 输出:
	//   - RawRuleSet: 原始 YAML、版本、摘要、来源。
	//
	// 错误:
	//   - 源不存在、IO 错误、摘要计算失败、上下文取消。
	Load(ctx context.Context) (RawRuleSet, error)

	// Watch 监听规则变化，阻塞直到 ctx 取消。
	//
	// 输入:
	//   - ctx: 控制停止。
	//   - onChange: 变化回调，参数为新的原始规则快照。
	//
	// 输出:
	//   - 无；ctx 取消或监听失败时返回。
	//
	// 错误:
	//   - 监听失败、上下文取消。
	Watch(ctx context.Context, onChange func(RawRuleSet)) error
}
