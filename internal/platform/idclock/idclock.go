// Package idclock 定义时间与 ID 生成端口，便于测试注入。
package idclock

import "time"

// Clock 是时间源。
type Clock interface {
	// Now 返回当前时间。
	// 输入: 无。输出: 当前时间。错误: 无。
	Now() time.Time
}

// IDGenerator 生成请求 ID 与任务 ID。
type IDGenerator interface {
	// NewRequestID 生成请求 ID。
	// 输入: 无。输出: 唯一字符串。错误: 无。
	NewRequestID() string

	// NewJobID 生成任务 ID。
	// 输入: 无。输出: 唯一字符串。错误: 无。
	NewJobID() string
}
