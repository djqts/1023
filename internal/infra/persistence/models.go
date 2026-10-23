package persistence

import "time"

// RuleSetModel 对应表 rule_set，保存规则集版本元数据与原始 YAML 快照。
type RuleSetModel struct {
	ID          uint64     `json:"id" db:"id" gorm:"column:id;primaryKey;autoIncrement"`
	Version     string     `json:"version" db:"version" gorm:"column:version;size:64;not null;uniqueIndex:uk_rule_set_version"`
	Digest      string     `json:"digest" db:"digest" gorm:"column:digest;size:64;not null;index:idx_rule_set_digest"`
	Content     string     `json:"content" db:"content" gorm:"column:content;type:longtext;not null"`
	Source      string     `json:"source" db:"source" gorm:"column:source;size:32;not null"`
	Status      string     `json:"status" db:"status" gorm:"column:status;size:16;not null;index:idx_rule_set_status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at" gorm:"column:created_at;autoCreateTime"`
	ActivatedAt *time.Time `json:"activated_at,omitempty" db:"activated_at" gorm:"column:activated_at"`
}

// RecognitionBatchModel 对应表 recognition_batch，保存批次审计摘要。
type RecognitionBatchModel struct {
	ID           uint64    `json:"id" db:"id" gorm:"column:id;primaryKey;autoIncrement"`
	RequestID    string    `json:"request_id" db:"request_id" gorm:"column:request_id;size:64;not null;uniqueIndex:uk_recognition_batch_request"`
	ItemCount    int       `json:"item_count" db:"item_count" gorm:"column:item_count;not null"`
	UnknownCount int       `json:"unknown_count" db:"unknown_count" gorm:"column:unknown_count;not null"`
	LatencyMS    int64     `json:"latency_ms" db:"latency_ms" gorm:"column:latency_ms;not null"`
	RuleDigest   string    `json:"rule_digest" db:"rule_digest" gorm:"column:rule_digest;size:64;not null;index:idx_recognition_batch_rule"`
	CreatedAt    time.Time `json:"created_at" db:"created_at" gorm:"column:created_at;autoCreateTime;index:idx_recognition_batch_created"`
}

// RecognitionItemModel 对应表 recognition_item，保存单条识别明细（P1，默认关闭/采样）。
type RecognitionItemModel struct {
	ID         uint64    `json:"id" db:"id" gorm:"column:id;primaryKey;autoIncrement"`
	BatchID    uint64    `json:"batch_id" db:"batch_id" gorm:"column:batch_id;not null;index:idx_recognition_item_batch"`
	IPHash     string    `json:"ip_hash" db:"ip_hash" gorm:"column:ip_hash;size:64;not null"`
	Port       int       `json:"port" db:"port" gorm:"column:port;not null"`
	Protocol   string    `json:"protocol" db:"protocol" gorm:"column:protocol;size:16;not null"`
	Product    string    `json:"product" db:"product" gorm:"column:product;size:64;not null"`
	Version    string    `json:"version" db:"version" gorm:"column:version;size:64;not null"`
	OSHint     string    `json:"os_hint" db:"os_hint" gorm:"column:os_hint;size:64;not null"`
	Confidence float64   `json:"confidence" db:"confidence" gorm:"column:confidence;not null"`
	RuleID     string    `json:"rule_id" db:"rule_id" gorm:"column:rule_id;size:64;not null"`
	CreatedAt  time.Time `json:"created_at" db:"created_at" gorm:"column:created_at;autoCreateTime"`
}

// AsyncJobModel 对应表 async_job，保存异步任务状态（P1）。
type AsyncJobModel struct {
	ID             string    `json:"id" db:"id" gorm:"column:id;type:varchar(64);primaryKey"`
	IdempotencyKey string    `json:"idempotency_key" db:"idempotency_key" gorm:"column:idempotency_key;size:128;not null;uniqueIndex:uk_async_job_idem"`
	Status         string    `json:"status" db:"status" gorm:"column:status;size:16;not null;index:idx_async_job_status"`
	InputRef       string    `json:"input_ref" db:"input_ref" gorm:"column:input_ref;size:256;not null"`
	ResultRef      string    `json:"result_ref" db:"result_ref" gorm:"column:result_ref;size:256;not null"`
	ErrorCode      string    `json:"error_code" db:"error_code" gorm:"column:error_code;size:64;not null"`
	CreatedAt      time.Time `json:"created_at" db:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// OutboxEventModel 对应表 outbox_event，保存待发布事件（P1）。
type OutboxEventModel struct {
	ID            uint64     `json:"id" db:"id" gorm:"column:id;primaryKey;autoIncrement"`
	AggregateType string     `json:"aggregate_type" db:"aggregate_type" gorm:"column:aggregate_type;size:32;not null"`
	AggregateID   string     `json:"aggregate_id" db:"aggregate_id" gorm:"column:aggregate_id;size:64;not null;index:idx_outbox_aggregate"`
	EventType     string     `json:"event_type" db:"event_type" gorm:"column:event_type;size:32;not null"`
	Payload       []byte     `json:"payload" db:"payload" gorm:"column:payload;type:json;not null"`
	Status        string     `json:"status" db:"status" gorm:"column:status;size:16;not null;index:idx_outbox_status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at" gorm:"column:created_at;autoCreateTime;index:idx_outbox_created"`
	PublishedAt   *time.Time `json:"published_at,omitempty" db:"published_at" gorm:"column:published_at"`
}
