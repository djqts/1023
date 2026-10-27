package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"banner-fingerprint/internal/engine"
	"banner-fingerprint/pkg/api"
)

func newTestServer(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	eng, err := engine.Load(filepath.Join("..", "..", "rules"))
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	return New(eng, Options{Version: "test"}).Router()
}

func TestFingerprintBareArrayAndHexEscapes(t *testing.T) {
	router := newTestServer(t)
	body := `[
		{"ip":"1.2.3.4","port":22,"banner":"SSH-2.0-OpenSSH_8.9p1 Ubuntu-3"},
		{"ip":"1.2.3.7","port":3306,"banner":"J\x00\x00\x00\n8.0.32\x00"},
		{"ip":"1.2.3.19","port":9999,"banner":"\x16\x03\x01\x00\xa5"}
	]`
	req := httptest.NewRequest(http.MethodPost, "/fingerprint", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "test-request")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("X-Request-ID"); got != "test-request" {
		t.Fatalf("request id = %q", got)
	}

	var results []api.FingerprintResult
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, rec.Body.String())
	}
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	if results[0].Protocol != "SSH" || results[0].Product != "OpenSSH" || results[0].Version != "8.9p1" || results[0].OSHint != "Ubuntu" {
		t.Fatalf("unexpected ssh result: %+v", results[0])
	}
	if results[1].Protocol != "MySQL" || results[1].Version != "8.0.32" {
		t.Fatalf("unexpected mysql result: %+v", results[1])
	}
	if results[2].Protocol != "unknown" || results[2].Confidence != 0 {
		t.Fatalf("unexpected unknown result: %+v", results[2])
	}
}

func TestFingerprintEnvelope(t *testing.T) {
	router := newTestServer(t)
	body := `{"items":[{"ip":"1.2.3.9","port":21,"banner":"220 ProFTPD 1.3.7 Server (ProFTPD)"}]}`
	req := httptest.NewRequest(http.MethodPost, "/fingerprint", strings.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var results []api.FingerprintResult
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(results) != 1 || results[0].Protocol != "FTP" || results[0].Product != "ProFTPD" {
		t.Fatalf("unexpected result: %+v", results)
	}
}

func TestFingerprintInvalidJSON(t *testing.T) {
	router := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/fingerprint", strings.NewReader(`{"items":`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHealth(t *testing.T) {
	router := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var health api.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &health); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if health.Status != api.StatusOK {
		t.Fatalf("status = %q", health.Status)
	}
	if health.Rules.Count == 0 || health.Rules.Digest == "" {
		t.Fatalf("rules health incomplete: %+v", health.Rules)
	}
}
