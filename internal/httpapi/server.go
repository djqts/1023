package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	"banner-fingerprint/internal/domain/fingerprint"
	"banner-fingerprint/internal/engine"
	"banner-fingerprint/pkg/api"
)

type requestIDContextKey struct{}

// Dependency 是健康检查中的一个外部依赖。
type Dependency struct {
	Name  string
	Check func(context.Context) error
}

// BatchAudit 是一次批量识别的审计摘要。
type BatchAudit struct {
	RequestID  string
	Total      int
	Unknown    int
	LatencyMS  int64
	RuleDigest string
	At         time.Time
}

// Options 是 HTTP 服务的可选配置。
type Options struct {
	Version        string
	MaxItems       int
	MaxBannerBytes int
	MaxBodyBytes   int64
	Logger         *slog.Logger
	Dependencies   []Dependency
	OnBatch        func(BatchAudit)
}

// Server 是 Gin HTTP 接入层。
type Server struct {
	engine  atomic.Pointer[engine.Engine]
	opts    Options
	logger  *slog.Logger
	started time.Time
}

// New 构造 HTTP 服务。
func New(eng *engine.Engine, opts Options) *Server {
	if opts.Version == "" {
		opts.Version = "dev"
	}
	if opts.MaxItems <= 0 {
		opts.MaxItems = 10000
	}
	if opts.MaxBannerBytes <= 0 {
		opts.MaxBannerBytes = 16 * 1024
	}
	if opts.MaxBodyBytes <= 0 {
		opts.MaxBodyBytes = 8 << 20
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	s := &Server{
		opts:    opts,
		logger:  opts.Logger,
		started: time.Now(),
	}
	s.engine.Store(eng)
	return s
}

// Reload 原子替换规则快照；在途请求继续使用旧快照。
func (s *Server) Reload(eng *engine.Engine) {
	if eng != nil {
		s.engine.Store(eng)
	}
}

// Router 返回注册了 /fingerprint 与 /health 的 Gin 引擎。
func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(s.requestIDMiddleware(), s.accessLogMiddleware())
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		s.logger.ErrorContext(c.Request.Context(), "panic recovered",
			"request_id", requestID(c),
			"panic", fmt.Sprint(recovered),
		)
		s.fail(c, http.StatusInternalServerError, "internal_error", "internal server error")
	}))

	r.POST("/fingerprint", s.handleFingerprint)
	r.GET("/health", s.handleHealth)

	r.NoMethod(func(c *gin.Context) {
		s.fail(c, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	})
	r.NoRoute(func(c *gin.Context) {
		s.fail(c, http.StatusNotFound, "not_found", "route not found")
	})
	return r
}

func (s *Server) handleFingerprint(c *gin.Context) {
	body, err := s.readBody(c)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			s.fail(c, http.StatusRequestEntityTooLarge, "payload_too_large", "request body too large")
			return
		}
		s.fail(c, http.StatusBadRequest, "invalid_request", "cannot read request body")
		return
	}

	items, err := api.DecodeBatch(body)
	if err != nil {
		s.fail(c, http.StatusBadRequest, "invalid_request", "invalid JSON batch")
		return
	}
	if len(items) > s.opts.MaxItems {
		s.fail(c, http.StatusRequestEntityTooLarge, "payload_too_large", "too many items")
		return
	}
	for _, item := range items {
		if len(item.Banner) > s.opts.MaxBannerBytes {
			s.fail(c, http.StatusRequestEntityTooLarge, "payload_too_large", "banner too large")
			return
		}
		if item.BannerB64 != "" {
			decoded, err := base64.StdEncoding.DecodeString(item.BannerB64)
			if err == nil && len(decoded) > s.opts.MaxBannerBytes {
				s.fail(c, http.StatusRequestEntityTooLarge, "payload_too_large", "banner too large")
				return
			}
		}
	}

	eng := s.engine.Load()
	if eng == nil || eng.RuleCount() == 0 {
		s.fail(c, http.StatusServiceUnavailable, "not_ready", "rules not loaded")
		return
	}

	start := time.Now()
	results := make([]api.FingerprintResult, 0, len(items))
	unknown := 0
	for _, item := range items {
		result := s.recognize(c.Request.Context(), item, eng)
		if result.Protocol == api.ProtocolUnknown {
			unknown++
		}
		results = append(results, result)
	}

	if s.opts.OnBatch != nil && len(items) > 0 {
		s.opts.OnBatch(BatchAudit{
			RequestID:  requestID(c),
			Total:      len(items),
			Unknown:    unknown,
			LatencyMS:  time.Since(start).Milliseconds(),
			RuleDigest: eng.Digest(),
			At:         time.Now(),
		})
	}

	c.JSON(http.StatusOK, results)
}

func (s *Server) recognize(ctx context.Context, item api.FingerprintItem, eng *engine.Engine) (result api.FingerprintResult) {
	defer func() {
		if recovered := recover(); recovered != nil {
			s.logger.ErrorContext(ctx, "fingerprint panic isolated",
				"request_id", requestIDFromContext(ctx),
				"panic", fmt.Sprint(recovered),
			)
			result = unknownResult(item)
		}
	}()

	banner, err := bannerBytes(item)
	if err != nil || len(banner) == 0 {
		return unknownResult(item)
	}

	out, err := eng.Fingerprint(fingerprint.Input{
		IP:     item.IP,
		Port:   item.Port,
		Banner: banner,
	})
	if err != nil {
		s.logger.WarnContext(ctx, "fingerprint failed",
			"request_id", requestIDFromContext(ctx),
			"error", err,
		)
		return unknownResult(item)
	}

	return api.FingerprintResult{
		IP:         out.IP,
		Port:       out.Port,
		Protocol:   string(out.Protocol),
		Product:    out.Product,
		Version:    out.Version,
		OSHint:     out.OSHint,
		Confidence: out.Confidence,
	}
}

func (s *Server) handleHealth(c *gin.Context) {
	eng := s.engine.Load()
	if eng == nil || eng.RuleCount() == 0 {
		c.JSON(http.StatusServiceUnavailable, api.HealthResponse{
			Status:  api.StatusNotReady,
			Version: s.opts.Version,
			Rules: api.RuleHealth{
				Count:    0,
				LoadedAt: s.started,
			},
			Dependencies:  map[string]api.DependencyHealth{},
			UptimeSeconds: time.Since(s.started).Seconds(),
		})
		return
	}

	status := api.StatusOK
	deps := make(map[string]api.DependencyHealth, len(s.opts.Dependencies))
	for _, dep := range s.opts.Dependencies {
		start := time.Now()
		checkCtx, cancel := context.WithTimeout(c.Request.Context(), 800*time.Millisecond)
		err := dep.Check(checkCtx)
		cancel()

		item := api.DependencyHealth{
			Status:    api.StatusOK,
			LatencyMS: time.Since(start).Milliseconds(),
		}
		if err != nil {
			item.Status = api.StatusDegraded
			item.Error = err.Error()
			status = api.StatusDegraded
		}
		deps[dep.Name] = item
	}

	c.JSON(http.StatusOK, api.HealthResponse{
		Status:  status,
		Version: s.opts.Version,
		Rules: api.RuleHealth{
			Count:    eng.RuleCount(),
			Digest:   eng.Digest(),
			Version:  "1",
			LoadedAt: s.started,
		},
		Dependencies:  deps,
		UptimeSeconds: time.Since(s.started).Seconds(),
	})
}

func (s *Server) readBody(c *gin.Context) ([]byte, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, s.opts.MaxBodyBytes)
	defer c.Request.Body.Close()
	return io.ReadAll(c.Request.Body)
}

func (s *Server) fail(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, api.ErrorResponse{
		Error: api.ErrorBody{
			Code:      code,
			Message:   message,
			RequestID: requestID(c),
		},
	})
}

func (s *Server) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := strings.TrimSpace(c.GetHeader("X-Request-ID"))
		if id == "" {
			id = newRequestID()
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		ctx := context.WithValue(c.Request.Context(), requestIDContextKey{}, id)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func (s *Server) accessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		s.logger.InfoContext(c.Request.Context(), "http_request",
			"request_id", requestID(c),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
	}
}

func requestID(c *gin.Context) string {
	return c.GetString("request_id")
}

func requestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDContextKey{}).(string); ok {
		return id
	}
	return ""
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func bannerBytes(item api.FingerprintItem) ([]byte, error) {
	if strings.TrimSpace(item.BannerB64) != "" {
		return base64.StdEncoding.DecodeString(item.BannerB64)
	}
	return []byte(item.Banner), nil
}

func unknownResult(item api.FingerprintItem) api.FingerprintResult {
	return api.FingerprintResult{
		IP:         item.IP,
		Port:       item.Port,
		Protocol:   api.ProtocolUnknown,
		Product:    "",
		Version:    "",
		OSHint:     "",
		Confidence: 0,
	}
}
