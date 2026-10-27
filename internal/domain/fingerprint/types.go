// Package fingerprint 定义识别领域的值对象、规则模型与核心端口。
//
// 本包不依赖 Gin、GORM、Redis 或任何传输/存储框架。
package fingerprint

import "time"

// Protocol 是协议标识。
type Protocol string

// 协议取值。unknown 表示无法识别，不是错误。
const (
	ProtocolUnknown Protocol = "unknown"
	ProtocolSSH     Protocol = "SSH"
	ProtocolHTTP    Protocol = "HTTP"
	ProtocolMySQL   Protocol = "MySQL"
	ProtocolRedis   Protocol = "Redis"
	ProtocolFTP     Protocol = "FTP"
)

// Input 是进入领域层的一条扫描数据。
type Input struct {
	IP     string `json:"ip" db:"-"`
	Port   int    `json:"port" db:"-"`
	Banner []byte `json:"banner,omitempty" db:"-"`
}

// Result 是领域层的识别结果，对应输出七字段。
type Result struct {
	IP         string   `json:"ip" db:"-"`
	Port       int      `json:"port" db:"-"`
	Protocol   Protocol `json:"protocol" db:"-"`
	Product    string   `json:"product" db:"-"`
	Version    string   `json:"version" db:"-"`
	OSHint     string   `json:"os_hint" db:"-"`
	Confidence float64  `json:"confidence" db:"-"`
}

// NormalizedBanner 是归一化后的横幅视图。
type NormalizedBanner struct {
	Text      string `json:"text" db:"-"`
	Bytes     []byte `json:"bytes" db:"-"`
	RawLength int    `json:"raw_length" db:"-"`
	Truncated bool   `json:"truncated" db:"-"`
	Binary    bool   `json:"binary" db:"-"`
}

// Evidence 是一条规则命中证据。
type Evidence struct {
	RuleID  string  `json:"rule_id" db:"-"`
	Kind    string  `json:"kind" db:"-"`
	Matched string  `json:"matched" db:"-"`
	Weight  float64 `json:"weight" db:"-"`
}

// ProtocolMatch 是协议判定结果。
type ProtocolMatch struct {
	Protocol   Protocol   `json:"protocol" db:"-"`
	Confidence float64    `json:"confidence" db:"-"`
	RuleID     string     `json:"rule_id" db:"-"`
	Evidence   []Evidence `json:"evidence,omitempty" db:"-"`
	Matched    bool       `json:"matched" db:"-"`
}

// ProductMatch 是产品与版本抽取结果。
type ProductMatch struct {
	Product    string  `json:"product" db:"-"`
	Version    string  `json:"version" db:"-"`
	Confidence float64 `json:"confidence" db:"-"`
	RuleID     string  `json:"rule_id,omitempty" db:"-"`
	Matched    bool    `json:"matched" db:"-"`
}

// OSHint 是 OS 线索。
type OSHint struct {
	Value      string  `json:"value" db:"-"`
	Confidence float64 `json:"confidence" db:"-"`
	RuleID     string  `json:"rule_id,omitempty" db:"-"`
	Matched    bool    `json:"matched" db:"-"`
}

// RuleStage 表示规则属于协议判定还是产品抽取阶段。
type RuleStage string

// 规则阶段取值。
const (
	StageProtocol RuleStage = "protocol"
	StageProduct  RuleStage = "product"
)

// RuleSet 是一次加载得到的不可变规则集。
type RuleSet struct {
	Version  string    `json:"version" db:"-"`
	Digest   string    `json:"digest" db:"-"`
	Source   string    `json:"source" db:"-"`
	LoadedAt time.Time `json:"loaded_at" db:"-"`
	Rules    []Rule    `json:"rules" db:"-"`
}

// Rule 是一条声明式识别规则。
type Rule struct {
	ID              string        `json:"id" db:"-"`
	Stage           RuleStage     `json:"stage" db:"-"`
	Protocol        Protocol      `json:"protocol" db:"-"`
	Priority        int           `json:"priority" db:"-"`
	Ports           []int         `json:"ports,omitempty" db:"-"`
	Match           MatchExpr     `json:"match" db:"-"`
	Product         string        `json:"product,omitempty" db:"-"`
	VersionTemplate string        `json:"version_template,omitempty" db:"-"`
	OSHintTemplate  string        `json:"os_hint_template,omitempty" db:"-"`
	Confidence      float64       `json:"confidence" db:"-"`
	Examples        []RuleExample `json:"examples,omitempty" db:"-"`
}

// MatchExpr 是规则匹配表达式，All 与 Any 可组合。
type MatchExpr struct {
	All []MatchTerm `json:"all,omitempty" db:"-"`
	Any []MatchTerm `json:"any,omitempty" db:"-"`
}

// MatchTerm 是单个匹配条件。
type MatchTerm struct {
	Prefix   string `json:"prefix,omitempty" db:"-"`
	Contains string `json:"contains,omitempty" db:"-"`
	Regex    string `json:"regex,omitempty" db:"-"`
	Offset   int    `json:"offset,omitempty" db:"-"`
	Encoding string `json:"encoding,omitempty" db:"-"`
	Capture  string `json:"capture,omitempty" db:"-"`
}

// RuleExample 是规则自带的自检样例。
type RuleExample struct {
	Banner   string   `json:"banner" db:"-"`
	Port     int      `json:"port" db:"-"`
	Protocol Protocol `json:"protocol" db:"-"`
	Product  string   `json:"product,omitempty" db:"-"`
	Version  string   `json:"version,omitempty" db:"-"`
	OSHint   string   `json:"os_hint,omitempty" db:"-"`
}
