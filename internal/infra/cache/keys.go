// Package cache 定义 Redis 键与缓存值约定。
package cache

// Redis 键前缀。具体拼接由适配器实现，键中禁止出现 banner 原文或原始 IP。
const (
	// KeyRuleActive 保存 active 规则快照（version、digest、原始 YAML）。
	KeyRuleActive = "fp:rule:active"

	// KeyResultPrefix 是结果缓存前缀，完整键为 fp:result:{digest}:{hash}。
	KeyResultPrefix = "fp:result:"

	// KeyRateLimitPrefix 是限流键前缀，完整键为 fp:ratelimit:{scope}:{id}:{window}。
	KeyRateLimitPrefix = "fp:ratelimit:"

	// KeyIdempotencyPrefix 是幂等键前缀，完整键为 fp:idem:{key}。
	KeyIdempotencyPrefix = "fp:idem:"

	// KeyJobPrefix 是任务状态缓存前缀，完整键为 fp:job:{id}。
	KeyJobPrefix = "fp:job:"

	// KeyRuleChangedChannel 是规则变更 Pub/Sub 频道。
	KeyRuleChangedChannel = "fp:rules:changed"

	// KeyLockPrefix 是 best-effort 协调锁前缀，完整键为 fp:lock:{name}。
	KeyLockPrefix = "fp:lock:"
)
