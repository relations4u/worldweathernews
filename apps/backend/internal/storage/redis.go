package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/relations4u/worldweathernews/apps/backend/internal/config"
)

// NewRedis öffnet einen Redis-Client per ParseURL.
// Caller schließt mit client.Close().
func NewRedis(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("initial ping: %w", err)
	}
	return client, nil
}

// RedisHealth prüft mit PING und 2s-Timeout.
func RedisHealth(ctx context.Context, client *redis.Client) error {
	if client == nil {
		return fmt.Errorf("client not initialized")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return client.Ping(ctx).Err()
}
