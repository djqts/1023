package messaging

import "time"

// Message 是 RabbitMQ 消息的通用信封。
type Message struct {
	ID          string            `json:"id" db:"-"`
	Exchange    string            `json:"exchange" db:"-"`
	RoutingKey  string            `json:"routing_key" db:"-"`
	Queue       string            `json:"queue" db:"-"`
	Body        []byte            `json:"body" db:"-"`
	Headers     map[string]string `json:"headers,omitempty" db:"-"`
	Timestamp   time.Time         `json:"timestamp" db:"-"`
	Redelivered bool              `json:"redelivered" db:"-"`
	DeliveryTag uint64            `json:"delivery_tag" db:"-"`
}
