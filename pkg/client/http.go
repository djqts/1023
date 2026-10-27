package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"banner-fingerprint/pkg/api"
)

// HTTPError 是 server 返回的非 2xx 错误。
type HTTPError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *HTTPError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("server %d %s: %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("server %d: %s", e.StatusCode, e.Message)
}

// HTTPClient 是 Client 的 HTTP 实现。
type HTTPClient struct {
	cfg  Config
	http *http.Client
}

var _ Client = (*HTTPClient)(nil)

// NewHTTP 构造 HTTP 客户端。
func NewHTTP(cfg Config) *HTTPClient {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = "http://127.0.0.1:8080"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 500
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "banner-fingerprint-client/1.0"
	}
	return &HTTPClient{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Fingerprint 发送批量识别请求。
func (c *HTTPClient) Fingerprint(ctx context.Context, req api.BatchRequest) (api.BatchResponse, error) {
	if req.Items == nil {
		req.Items = []api.FingerprintItem{}
	}
	payload, err := json.Marshal(req)
	if err != nil {
		return api.BatchResponse{}, fmt.Errorf("encode request: %w", err)
	}
	body, err := c.do(ctx, http.MethodPost, "/fingerprint", payload)
	if err != nil {
		return api.BatchResponse{}, err
	}
	results, err := api.DecodeResults(body)
	if err != nil {
		return api.BatchResponse{}, fmt.Errorf("decode response: %w", err)
	}
	return api.BatchResponse{Results: results}, nil
}

// Health 查询健康状态。
func (c *HTTPClient) Health(ctx context.Context) (api.HealthResponse, error) {
	body, err := c.do(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		return api.HealthResponse{}, err
	}
	var resp api.HealthResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return api.HealthResponse{}, fmt.Errorf("decode health: %w", err)
	}
	return resp, nil
}

func (c *HTTPClient) do(ctx context.Context, method, path string, payload []byte) ([]byte, error) {
	url := strings.TrimRight(c.cfg.BaseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var er api.ErrorResponse
		if json.Unmarshal(body, &er) == nil && er.Error.Message != "" {
			return nil, &HTTPError{StatusCode: resp.StatusCode, Code: er.Error.Code, Message: er.Error.Message}
		}
		return nil, &HTTPError{StatusCode: resp.StatusCode, Message: strings.TrimSpace(string(body))}
	}
	return body, nil
}
