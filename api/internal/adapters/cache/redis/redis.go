package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"atom-maintenance/internal/config"
	"atom-maintenance/internal/domain"
	"atom-maintenance/internal/ports"
	"atom-maintenance/platform/logger"

	rds "github.com/redis/go-redis/v9"
)

type redisAdapter struct {
	client  *rds.Client
	log     *slog.Logger
	cacheTO time.Duration
}

func New(cfg config.RedisConfig, log *slog.Logger) (ports.Cache, *rds.Client, error) {
	opt, err := rds.ParseURL(cfg.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("parse redis url: %w", err)
	}

	opt.PoolSize = cfg.PoolSize
	opt.MinIdleConns = cfg.MinIdleConns
	opt.DialTimeout = cfg.DialTimeout

	client := rds.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, nil, fmt.Errorf("ping redis: %w", err)
	}

	return &redisAdapter{client: client, log: log, cacheTO: cfg.DialTimeout}, client, nil
}

func (a *redisAdapter) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	log := logger.WithReqID(ctx, a.log).With("module", "redis", "op", "Set")
	start := time.Now()

	cacheCtx, cancel := context.WithTimeout(ctx, a.cacheTO)
	defer cancel()

	if err := a.client.Set(cacheCtx, key, value, ttl).Err(); err != nil {
		log.Error("set failed", "key", key, "err", err, "duration_ms", time.Since(start).Milliseconds())
		return err
	}

	log.Info("key set", "key", key, "ttl", ttl.String(), "duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (a *redisAdapter) Get(ctx context.Context, key string) (string, error) {
	log := logger.WithReqID(ctx, a.log).With("module", "redis", "op", "Get")
	start := time.Now()

	cacheCtx, cancel := context.WithTimeout(ctx, a.cacheTO)
	defer cancel()

	val, err := a.client.Get(cacheCtx, key).Result()
	if err == rds.Nil {
		log.Warn("cache miss", "key", key, "duration_ms", time.Since(start).Milliseconds())
		return "", domain.ErrCacheMiss
	}
	if err != nil {
		log.Error("get failed", "key", key, "err", err, "duration_ms", time.Since(start).Milliseconds())
		return "", err
	}

	log.Info("cache hit", "key", key, "duration_ms", time.Since(start).Milliseconds())
	return val, nil
}

func (a *redisAdapter) Delete(ctx context.Context, key string) error {
	log := logger.WithReqID(ctx, a.log).With("module", "redis", "op", "Delete")
	start := time.Now()

	cacheCtx, cancel := context.WithTimeout(ctx, a.cacheTO)
	defer cancel()

	if err := a.client.Del(cacheCtx, key).Err(); err != nil {
		log.Error("delete failed", "key", key, "err", err, "duration_ms", time.Since(start).Milliseconds())
		return err
	}

	log.Info("key deleted", "key", key, "duration_ms", time.Since(start).Milliseconds())
	return nil
}
