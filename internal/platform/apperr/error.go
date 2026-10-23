// Package apperr 定义统一错误码与错误数据载体。
//
// Error 的 Error()/Unwrap() 方法在实现阶段补齐；本包只保留契约。
package apperr

// Code 是稳定的机器可读错误码。
type Code string

// 错误码取值。
const (
	CodeInvalidRequest   Code = "invalid_request"
	CodePayloadTooLarge  Code = "payload_too_large"
	CodeUnsupportedMedia Code = "unsupported_media_type"
	CodeNotFound         Code = "not_found"
	CodeConflict         Code = "conflict"
	CodeInternal         Code = "internal_error"
	CodeUnavailable      Code = "unavailable"
	CodeNotReady         Code = "not_ready"
	CodeOverloaded       Code = "overloaded"
)

// CodedError 是可映射到 HTTP/gRPC 的错误契约。
type CodedError interface {
	error
	Code() Code
	PublicMessage() string
}

// Error 是错误数据载体。
type Error struct {
	Code      Code   `json:"code" db:"-"`
	Message   string `json:"message" db:"-"`
	RequestID string `json:"request_id,omitempty" db:"-"`
	Op        string `json:"op,omitempty" db:"-"`
	Retryable bool   `json:"retryable" db:"-"`
	Cause     error  `json:"-" db:"-"`
}
