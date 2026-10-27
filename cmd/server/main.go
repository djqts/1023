package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"banner-fingerprint/internal/engine"
	"banner-fingerprint/internal/httpapi"
	"banner-fingerprint/internal/infra/persistence"
	"banner-fingerprint/pkg/api"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(runHealthcheck(os.Args[2:]))
	}
	os.Exit(runServer())
}

func runHealthcheck(args []string) int {
	fs := flag.NewFlagSet("healthcheck", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	url := fs.String("url", "http://127.0.0.1:8080/health", "health endpoint")
	timeout := fs.Duration("timeout", 3*time.Second, "request timeout")
	if err := fs.Parse(args); err != nil {
		return 1
	}

	client := &http.Client{Timeout: *timeout}
	resp, err := client.Get(*url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck request: %v\n", err)
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "healthcheck status: %d\n", resp.StatusCode)
		return 1
	}

	var health api.HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck decode: %v\n", err)
		return 1
	}
	if health.Status != api.StatusOK && health.Status != api.StatusDegraded {
		return 1
	}
	return 0
}

func runServer() int {
	logger := newLogger()
	slog.SetDefault(logger)

	rulesDir := envString("RULES_DIR", "rules")
	eng, err := engine.Load(rulesDir)
	if err != nil {
		logger.Error("load rules failed", "rules_dir", rulesDir, "error", err)
		return 1
	}
	logger.Info("rules loaded", "count", eng.RuleCount(), "digest", eng.Digest(), "rules_dir", rulesDir)

	profile := strings.ToLower(envString("PROFILE", "p0b"))
	var (
		db      *gorm.DB
		sqlDB   *sql.DB
		rdb     *redis.Client
		auditCh chan persistence.RecognitionBatchModel
	)
	if profile != "p0a" {
		db, sqlDB = initMySQL(logger)
		rdb = initRedis(logger)
		if db != nil {
			auditCh = startAuditWorker(logger, db)
		}
	}

	deps := make([]httpapi.Dependency, 0, 2)
	if sqlDB != nil {
		deps = append(deps, httpapi.Dependency{
			Name: "mysql",
			Check: func(ctx context.Context) error {
				return sqlDB.PingContext(ctx)
			},
		})
	}
	if rdb != nil {
		deps = append(deps, httpapi.Dependency{
			Name: "redis",
			Check: func(ctx context.Context) error {
				return rdb.Ping(ctx).Err()
			},
		})
	}

	var onBatch func(httpapi.BatchAudit)
	if auditCh != nil {
		onBatch = func(audit httpapi.BatchAudit) {
			select {
			case auditCh <- persistence.RecognitionBatchModel{
				RequestID:    audit.RequestID,
				ItemCount:    audit.Total,
				UnknownCount: audit.Unknown,
				LatencyMS:    audit.LatencyMS,
				RuleDigest:   audit.RuleDigest,
				CreatedAt:    audit.At,
			}:
			default:
				logger.Warn("audit queue full, dropping summary", "request_id", audit.RequestID)
			}
		}
	}

	handler := httpapi.New(eng, httpapi.Options{
		Version:        envString("APP_VERSION", "dev"),
		MaxItems:       envInt("MAX_ITEMS", 10000),
		MaxBannerBytes: envInt("MAX_BANNER_BYTES", 16*1024),
		MaxBodyBytes:   envInt64("MAX_BODY_BYTES", 8<<20),
		Logger:         logger,
		Dependencies:   deps,
		OnBatch:        onBatch,
	})

	gin.SetMode(gin.ReleaseMode)
	router := handler.Router()
	httpServer := &http.Server{
		Addr:              envString("HTTP_ADDR", ":8080"),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go watchRuleReload(ctx, func() error {
		next, err := engine.Load(rulesDir)
		if err != nil {
			return err
		}
		handler.Reload(next)
		logger.Info("rules reloaded", "count", next.RuleCount(), "digest", next.Digest())
		return nil
	}, logger)

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", httpServer.Addr, "profile", profile)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			return 1
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			return 1
		}
	}

	if rdb != nil {
		_ = rdb.Close()
	}
	if sqlDB != nil {
		_ = sqlDB.Close()
	}
	return 0
}

func newLogger() *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

func initMySQL(logger *slog.Logger) (*gorm.DB, *sql.DB) {
	dsn := strings.TrimSpace(os.Getenv("MYSQL_DSN"))
	if dsn == "" {
		return nil, nil
	}
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{Logger: gormlogger.Discard})
	if err != nil {
		logger.Warn("mysql disabled: open failed", "error", err)
		return nil, nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.Warn("mysql disabled: get sql.DB failed", "error", err)
		return nil, nil
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		logger.Warn("mysql disabled: ping failed", "error", err)
		_ = sqlDB.Close()
		return nil, nil
	}
	logger.Info("mysql connected")
	return db, sqlDB
}

func initRedis(logger *slog.Logger) *redis.Client {
	addr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if addr == "" {
		return nil
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DialTimeout:  2 * time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("redis disabled: ping failed", "error", err)
		_ = rdb.Close()
		return nil
	}
	logger.Info("redis connected")
	return rdb
}

func startAuditWorker(logger *slog.Logger, db *gorm.DB) chan persistence.RecognitionBatchModel {
	ch := make(chan persistence.RecognitionBatchModel, 256)
	go func() {
		for model := range ch {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err := db.WithContext(ctx).Table("recognition_batch").Create(&model).Error
			cancel()
			if err != nil {
				logger.Warn("audit insert failed", "request_id", model.RequestID, "error", err)
			}
		}
	}()
	return ch
}

func envString(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envInt64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
