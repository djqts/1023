package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"banner-fingerprint/pkg/api"
)

func TestFingerprintChunksAndRetries(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := atomic.AddInt32(&calls, 1)
		var req api.BatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if call == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		results := make([]api.FingerprintResult, len(req.Items))
		for i, item := range req.Items {
			results[i] = api.FingerprintResult{
				IP:         item.IP,
				Port:       item.Port,
				Protocol:   api.ProtocolUnknown,
				Confidence: 0,
			}
		}
		_ = json.NewEncoder(w).Encode(results)
	}))
	defer server.Close()

	cli := NewHTTP(Config{
		BaseURL:      server.URL,
		BatchSize:    2,
		MaxRetries:   1,
		RetryBackoff: time.Millisecond,
	})
	items := []api.FingerprintItem{
		{IP: "1.1.1.1", Port: 1},
		{IP: "1.1.1.2", Port: 2},
		{IP: "1.1.1.3", Port: 3},
	}
	resp, err := cli.Fingerprint(context.Background(), api.BatchRequest{Items: items})
	if err != nil {
		t.Fatalf("fingerprint: %v", err)
	}
	if len(resp.Results) != 3 {
		t.Fatalf("got %d results, want 3", len(resp.Results))
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("got %d calls, want 3 (2-item chunk retried once + 1-item chunk)", got)
	}
}

func TestFingerprintDoesNotRetry400(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"invalid_request","message":"bad"}}`))
	}))
	defer server.Close()

	cli := NewHTTP(Config{
		BaseURL:      server.URL,
		BatchSize:    10,
		MaxRetries:   3,
		RetryBackoff: time.Millisecond,
	})
	_, err := cli.Fingerprint(context.Background(), api.BatchRequest{Items: []api.FingerprintItem{{IP: "1.1.1.1", Port: 1}}})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("got %d calls, want 1", got)
	}
}
