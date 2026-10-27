package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"banner-fingerprint/internal/engine"
)

func TestReloadSwapsSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng, err := engine.Load(filepath.Join("..", "..", "rules"))
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	server := New(eng, Options{Version: "test"})
	router := server.Router()

	first := healthDigest(t, router)
	if first == "" {
		t.Fatal("empty digest")
	}

	dir, err := os.MkdirTemp(".", ".reload-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	copyRules(t, filepath.Join("..", "..", "rules"), dir)
	extra := []byte("version: \"1\"\nrules:\n  - id: reload-extra\n    protocol: SMTP\n    product: ReloadTest\n    priority: 1\n    confidence: 0.5\n    patterns:\n      - '^RELOAD-TEST'\n")
	if err := os.WriteFile(filepath.Join(dir, "99-extra.yaml"), extra, 0o644); err != nil {
		t.Fatal(err)
	}

	next, err := engine.Load(dir)
	if err != nil {
		t.Fatalf("load updated rules: %v", err)
	}
	server.Reload(next)

	second := healthDigest(t, router)
	if second == first {
		t.Fatal("digest did not change after reload")
	}
}

func healthDigest(t *testing.T, router http.Handler) string {
	t.Helper()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("health status = %d", rec.Code)
	}
	var health struct {
		Rules struct {
			Digest string `json:"digest"`
		} `json:"rules"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &health); err != nil {
		t.Fatal(err)
	}
	return health.Rules.Digest
}

func copyRules(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dst, entry.Name()), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
