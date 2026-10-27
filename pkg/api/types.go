// Package api 定义 client 与 server 共享的 HTTP 数据传输对象。
//
// 这些结构只描述契约，不包含校验、持久化或识别逻辑。
package api

import "time"

// 协议取值。unknown 是识别失败时的统一值，不是错误。
const (
	ProtocolUnknown = "unknown"
	ProtocolSSH     = "SSH"
	ProtocolHTTP    = "HTTP"
	ProtocolMySQL   = "MySQL"
	ProtocolRedis   = "Redis"
	ProtocolFTP     = "FTP"
	ProtocolSMTP    = "SMTP"
)

// 健康状态取值。
const (
	StatusOK       = "ok"
	StatusDegraded = "degraded"
	StatusNotReady = "not_ready"
)

// FingerprintItem 是一条待识别的扫描原始数据。
type FingerprintItem struct {
	IP        string `json:"ip" db:"-"`
	Port      int    `json:"port" db:"-"`
	Banner    string `json:"banner,omitempty" db:"-"`
	BannerB64 string `json:"banner_b64,omitempty" db:"-"`
}

// BatchRequest 是 POST /fingerprint 的请求体。
type BatchRequest struct {
	Items []FingerprintItem `json:"items" db:"-"`
}

// FingerprintResult 是一条识别结果，严格对应需求的七个字段。
type FingerprintResult struct {
	IP         string  `json:"ip" db:"-"`
	Port       int     `json:"port" db:"-"`
	Protocol   string  `json:"protocol" db:"-"`
	Product    string  `json:"product" db:"-"`
	Version    string  `json:"version" db:"-"`
	OSHint     string  `json:"os_hint" db:"-"`
	Confidence float64 `json:"confidence" db:"-"`
}

// BatchResponse 是 POST /fingerprint 的响应体。
type BatchResponse struct {
	Results []FingerprintResult `json:"results" db:"-"`
}

// RuleHealth 是健康检查中的规则集状态。
type RuleHealth struct {
	Count    int       `json:"count" db:"-"`
	Digest   string    `json:"digest" db:"-"`
	Version  string    `json:"version" db:"-"`
	LoadedAt time.Time `json:"loaded_at" db:"-"`
}

// DependencyHealth 是健康检查中的单个依赖状态。
type DependencyHealth struct {
	Status    string `json:"status" db:"-"`
	LatencyMS int64  `json:"latency_ms" db:"-"`
	Error     string `json:"error,omitempty" db:"-"`
}

// HealthResponse 是 GET /health 的响应体。
type HealthResponse struct {
	Status        string                      `json:"status" db:"-"`
	Version       string                      `json:"version" db:"-"`
	Rules         RuleHealth                  `json:"rules" db:"-"`
	Dependencies  map[string]DependencyHealth `json:"dependencies" db:"-"`
	UptimeSeconds float64                     `json:"uptime_s" db:"-"`
}

// ErrorBody 是统一错误体的内容。
type ErrorBody struct {
	Code      string `json:"code" db:"-"`
	Message   string `json:"message" db:"-"`
	RequestID string `json:"request_id,omitempty" db:"-"`
}

// ErrorResponse 是所有非 2xx 响应的统一外壳。
type ErrorResponse struct {
	Error ErrorBody `json:"error" db:"-"`
}
