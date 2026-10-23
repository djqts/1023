package fingerprint

import "context"

// Engine 是单条识别入口。
type Engine interface {
	// Fingerprint 识别一条扫描数据。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//   - in: 单条 ip/port/banner，调用方已完成结构校验。
	//
	// 输出:
	//   - Result: 七字段结果；无法识别时 Protocol=unknown、Confidence=0。
	//
	// 错误:
	//   - 仅内部错误（规则快照不可用、上下文取消等）。
	//   - 认不出不返回 error。
	Fingerprint(ctx context.Context, in Input) (Result, error)
}

// Normalizer 把原始字节转换为可匹配的文本/字节视图。
type Normalizer interface {
	// Normalize 归一化横幅。
	//
	// 输入:
	//   - raw: 原始 banner 字节，可能为空、截断或非 UTF-8。
	//
	// 输出:
	//   - NormalizedBanner: 文本视图、字节视图和异常标记。
	//
	// 错误:
	//   - 不返回 error；数据异常通过 Truncated/Binary 标记。
	Normalize(raw []byte) NormalizedBanner
}

// ProtocolClassifier 判定协议。
type ProtocolClassifier interface {
	// Classify 根据横幅与端口判定协议。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//   - banner: 归一化后的横幅。
	//   - port: 目标端口，仅作为先验/加权，不作为唯一判据。
	//
	// 输出:
	//   - ProtocolMatch: 协议、置信度、规则 ID、证据；无命中时 Matched=false。
	//
	// 错误:
	//   - 规则集损坏、上下文取消等内部错误。
	Classify(ctx context.Context, banner NormalizedBanner, port int) (ProtocolMatch, error)
}

// ProductExtractor 抽取产品与版本。
type ProductExtractor interface {
	// Extract 在已知协议下抽取产品与版本。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//   - banner: 归一化后的横幅。
	//   - protocol: 已判定的协议。
	//
	// 输出:
	//   - ProductMatch: 产品、版本、置信度；抽不到时字段留空、Matched=false。
	//
	// 错误:
	//   - 规则集损坏、上下文取消等内部错误。
	Extract(ctx context.Context, banner NormalizedBanner, protocol Protocol) (ProductMatch, error)
}

// OSExtractor 推断 OS 线索。
type OSExtractor interface {
	// Extract 根据横幅与产品推断 OS 线索。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//   - banner: 归一化后的横幅。
	//   - product: 已抽取的产品信息。
	//
	// 输出:
	//   - OSHint: OS 线索；无证据时 Value 为空、Matched=false。
	//
	// 错误:
	//   - 规则集损坏、上下文取消等内部错误。
	Extract(ctx context.Context, banner NormalizedBanner, product ProductMatch) (OSHint, error)
}

// ConfidenceCalculator 根据证据合成置信度。
type ConfidenceCalculator interface {
	// Calculate 计算 0~1 的置信度。
	//
	// 输入:
	//   - evidence: 规则命中证据列表。
	//
	// 输出:
	//   - float64: [0,1] 的确定性分数。
	//
	// 错误:
	//   - 不返回 error；无证据时返回 0。
	Calculate(evidence []Evidence) float64
}

// RuleProvider 提供当前不可变规则快照。
type RuleProvider interface {
	// Snapshot 返回当前生效的规则集。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//
	// 输出:
	//   - RuleSet: 不可变快照，调用方不得修改。
	//
	// 错误:
	//   - 规则未加载、规则损坏、上下文取消。
	Snapshot(ctx context.Context) (RuleSet, error)
}

// RuleValidator 校验规则集的结构与正则合法性。
type RuleValidator interface {
	// Validate 校验规则集。
	//
	// 输入:
	//   - ctx: 控制取消与超时。
	//   - set: 待校验的规则集。
	//
	// 输出:
	//   - 无；校验通过返回 nil。
	//
	// 错误:
	//   - 非法正则、重复规则 ID、缺失捕获组、置信度越界、schema 版本不兼容。
	Validate(ctx context.Context, set RuleSet) error
}
