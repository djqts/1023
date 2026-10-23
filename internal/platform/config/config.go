// Package config 定义服务配置结构；加载与校验逻辑在实现阶段补齐。
package config

import "time"

// Config 是 server 的完整配置。
type Config struct {
	App      AppConfig      `json:"app" db:"-"`
	HTTP     HTTPConfig     `json:"http" db:"-"`
	MySQL    MySQLConfig    `json:"mysql" db:"-"`
	Redis    RedisConfig    `json:"redis" db:"-"`
	RabbitMQ RabbitMQConfig `json:"rabbitmq" db:"-"`
	Rules    RulesConfig    `json:"rules" db:"-"`
	Limits   LimitsConfig   `json:"limits" db:"-"`
	Log      LogConfig      `json:"log" db:"-"`
	Profile  ProfileConfig  `json:"profile" db:"-"`
}

// AppConfig 是应用元信息与停机参数。
type AppConfig struct {
	Name            string        `json:"name" db:"-"`
	Env             string        `json:"env" db:"-"`
	Version         string        `json:"version" db:"-"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout" db:"-"`
}

// HTTPConfig 是 HTTP 服务参数。
type HTTPConfig struct {
	Addr              string        `json:"addr" db:"-"`
	ReadHeaderTimeout time.Duration `json:"read_header_timeout" db:"-"`
	ReadTimeout       time.Duration `json:"read_timeout" db:"-"`
	WriteTimeout      time.Duration `json:"write_timeout" db:"-"`
	IdleTimeout       time.Duration `json:"idle_timeout" db:"-"`
	MaxHeaderBytes    int           `json:"max_header_bytes" db:"-"`
}

// MySQLConfig 是 MySQL 连接与连接池参数。
type MySQLConfig struct {
	DSN             string        `json:"dsn" db:"-"`
	MaxOpenConns    int           `json:"max_open_conns" db:"-"`
	MaxIdleConns    int           `json:"max_idle_conns" db:"-"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" db:"-"`
	QueryTimeout    time.Duration `json:"query_timeout" db:"-"`
}

// RedisConfig 是 Redis 连接、缓存 TTL 与限流参数。
type RedisConfig struct {
	Addr            string        `json:"addr" db:"-"`
	Password        string        `json:"password" db:"-"`
	DB              int           `json:"db" db:"-"`
	PoolSize        int           `json:"pool_size" db:"-"`
	DialTimeout     time.Duration `json:"dial_timeout" db:"-"`
	ReadTimeout     time.Duration `json:"read_timeout" db:"-"`
	WriteTimeout    time.Duration `json:"write_timeout" db:"-"`
	RuleCacheTTL    time.Duration `json:"rule_cache_ttl" db:"-"`
	ResultCacheTTL  time.Duration `json:"result_cache_ttl" db:"-"`
	JobStatusTTL    time.Duration `json:"job_status_ttl" db:"-"`
	RateLimitWindow time.Duration `json:"rate_limit_window" db:"-"`
}

// RabbitMQConfig 是消息队列参数（P1）。
type RabbitMQConfig struct {
	Enabled        bool          `json:"enabled" db:"-"`
	URL            string        `json:"url" db:"-"`
	Exchange       string        `json:"exchange" db:"-"`
	JobQueue       string        `json:"job_queue" db:"-"`
	AuditQueue     string        `json:"audit_queue" db:"-"`
	Prefetch       int           `json:"prefetch" db:"-"`
	ReconnectDelay time.Duration `json:"reconnect_delay" db:"-"`
}

// RulesConfig 是规则源与热加载参数。
type RulesConfig struct {
	Dir            string        `json:"dir" db:"-"`
	WatchEnabled   bool          `json:"watch_enabled" db:"-"`
	ReloadInterval time.Duration `json:"reload_interval" db:"-"`
	CacheTTL       time.Duration `json:"cache_ttl" db:"-"`
	FailFast       bool          `json:"fail_fast" db:"-"`
}

// LimitsConfig 是请求边界。
type LimitsConfig struct {
	MaxItems       int           `json:"max_items" db:"-"`
	MaxBannerBytes int           `json:"max_banner_bytes" db:"-"`
	MaxBodyBytes   int64         `json:"max_body_bytes" db:"-"`
	RequestTimeout time.Duration `json:"request_timeout" db:"-"`
	MaxConcurrency int           `json:"max_concurrency" db:"-"`
}

// LogConfig 是日志参数。
type LogConfig struct {
	Level  string `json:"level" db:"-"`
	Format string `json:"format" db:"-"`
}

// ProfileConfig 选择 P0-A/P0-B/P1 运行形态。
type ProfileConfig struct {
	Name            string `json:"name" db:"-"`
	MySQLEnabled    bool   `json:"mysql_enabled" db:"-"`
	RedisEnabled    bool   `json:"redis_enabled" db:"-"`
	RabbitMQEnabled bool   `json:"rabbitmq_enabled" db:"-"`
	GRPCEnabled     bool   `json:"grpc_enabled" db:"-"`
}
