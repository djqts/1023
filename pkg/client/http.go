package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

// PartialError 表示部分批次成功、部分批次失败。
type PartialError struct {
	Completed int
	Total     int
	Err       error
}

func (e *PartialError) Error() string {
	return fmt.Sprintf("partial failure: %d/%d items completed: %v", e.Completed, e.Total, e.Err)
}

func (e *PartialError) Unwrap() error { return e.Err }

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
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = 200 * time.Millisecond
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "banner-fingerprint-client/1.0"
	}
	return &HTTPClient{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Fingerprint 分片发送批量识别请求，并对可重试错误做指数退避。
func (c *HTTPClient) Fingerprint(ctx context.Context, req api.BatchRequest) (api.BatchResponse, error) {
	if len(req.Items) == 0 {
		return api.BatchResponse{Results: []api.FingerprintResult{}}, nil
	}

	var all []api.FingerprintResult
	for start := 0; start < len(req.Items); start += c.cfg.BatchSize {
		end := start + c.cfg.BatchSize
		if end > len(req.Items) {
			end = len(req.Items)
		}
		part := req.Items[start:end]

		payload, err := json.Marshal(api.BatchRequest{Items: part})
		if err != nil {
			return api.BatchResponse{Results: all}, fmt.Errorf("encode request: %w", err)
		}
		body, err := c.doWithRetry(ctx, http.MethodPost, "/fingerprint", payload)
		if err != nil {
			if len(all) == 0 {
				return api.BatchResponse{}, err
			}
			return api.BatchResponse{Results: all}, &PartialError{Completed: len(all), Total: len(req.Items), Err: err}
		}
		results, err := api.DecodeResults(body)
		if err != nil {
			if len(all) == 0 {
				return api.BatchResponse{}, fmt.Errorf("decode response: %w", err)
			}
			return api.BatchResponse{Results: all}, &PartialError{Completed: len(all), Total: len(req.Items), Err: err}
		}
		if len(results) != len(part) {
			err = fmt.Errorf("batch result count %d != %d", len(results), len(part))
			if len(all) == 0 {
				return api.BatchResponse{}, err
			}
			return api.BatchResponse{Results: all}, &PartialError{Completed: len(all), Total: len(req.Items), Err: err}
		}
		all = append(all, results...)
	}
	return api.BatchResponse{Results: all}, nil
}

// Health 查询健康状态。
func (c *HTTPClient) Health(ctx context.Context) (api.HealthResponse, error) {
	body, err := c.doWithRetry(ctx, http.MethodGet, "/health", nil)
	if err != nil {
		return api.HealthResponse{}, err
	}
	var resp api.HealthResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return api.HealthResponse{}, fmt.Errorf("decode health: %w", err)
	}
	return resp, nil
}

func (c *HTTPClient) doWithRetry(ctx context.Context, method, path string, payload []byte) ([]byte, error) {
	attempts := c.cfg.MaxRetries + 1
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		body, err := c.do(ctx, method, path, payload)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if attempt == attempts-1 || !retryable(err) || ctx.Err() != nil {
			break
		}
		delay := c.cfg.RetryBackoff * time.Duration(1<<attempt)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
	return nil, lastErr
}

func retryable(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == http.StatusTooManyRequests || httpErr.StatusCode >= 500
	}
	return !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
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
