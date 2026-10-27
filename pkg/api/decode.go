package api

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// NormalizeHexEscapes 把非标准 JSON 的 \xHH 转成合法的 \u00HH。
//
// 扫描器导出数据里常见 "J\x00\x00\x00\n8.0.32\x00" 这种写法；
// encoding/json 只接受 \uXXXX，这里先做一次无损的 ASCII 控制字符转换。
func NormalizeHexEscapes(data []byte) []byte {
	out := make([]byte, 0, len(data))
	for i := 0; i < len(data); i++ {
		if data[i] == '\\' && i+3 < len(data) && data[i+1] == 'x' && isHex(data[i+2]) && isHex(data[i+3]) {
			out = append(out, '\\', 'u', '0', '0', data[i+2], data[i+3])
			i += 3
			continue
		}
		out = append(out, data[i])
	}
	return out
}

// DecodeBatch 解析批量请求，兼容裸数组与 {"items":[...]} 两种形状。
func DecodeBatch(data []byte) ([]FingerprintItem, error) {
	data = NormalizeHexEscapes(data)
	data = bytes.TrimSpace(bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}))
	if len(data) == 0 {
		return []FingerprintItem{}, nil
	}

	switch data[0] {
	case '[':
		var items []FingerprintItem
		if err := json.Unmarshal(data, &items); err != nil {
			return nil, fmt.Errorf("decode batch array: %w", err)
		}
		return nonNilItems(items), nil
	case '{':
		var req BatchRequest
		if err := json.Unmarshal(data, &req); err == nil && req.Items != nil {
			return nonNilItems(req.Items), nil
		}
		// 兼容单个 item 对象。
		var item FingerprintItem
		if err := json.Unmarshal(data, &item); err == nil && (item.IP != "" || item.Port != 0 || item.Banner != "" || item.BannerB64 != "") {
			return []FingerprintItem{item}, nil
		}
		return nil, fmt.Errorf("decode batch object: expected items array")
	default:
		return nil, fmt.Errorf("decode batch: expected array or object")
	}
}

// DecodeResults 解析响应，兼容裸数组与 {"results":[...]}。
func DecodeResults(data []byte) ([]FingerprintResult, error) {
	data = bytes.TrimSpace(bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF}))
	if len(data) == 0 {
		return []FingerprintResult{}, nil
	}

	switch data[0] {
	case '[':
		var results []FingerprintResult
		if err := json.Unmarshal(data, &results); err != nil {
			return nil, fmt.Errorf("decode results array: %w", err)
		}
		return nonNilResults(results), nil
	case '{':
		var resp BatchResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("decode results object: %w", err)
		}
		return nonNilResults(resp.Results), nil
	default:
		return nil, fmt.Errorf("decode results: expected array or object")
	}
}

func nonNilItems(items []FingerprintItem) []FingerprintItem {
	if items == nil {
		return []FingerprintItem{}
	}
	return items
}

func nonNilResults(results []FingerprintResult) []FingerprintResult {
	if results == nil {
		return []FingerprintResult{}
	}
	return results
}

func isHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}
