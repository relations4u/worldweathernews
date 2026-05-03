// Package storage kapselt die Verbindungs-Pools für PostgreSQL und Redis.
package storage

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/relations4u/worldweathernews/apps/backend/internal/config"
)

// clampInt32 schneidet einen int sicher auf den int32-Bereich zu — verhindert
// gosec G115 bei den (kleinen, konfigurierbaren) Pool-Limits.
func clampInt32(v int) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}

// NewPool öffnet einen pgxpool.Pool gemäß DatabaseConfig.
// Caller schließt mit pool.Close().
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	pcfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		pcfg.MaxConns = clampInt32(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		pcfg.MinConns = clampInt32(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		pcfg.MaxConnLifetime = cfg.ConnMaxLifetime
	}

	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("initial ping: %w", err)
	}

	return pool, nil
}

// PostgresHealth prüft die Verbindung mit `SELECT 1` und einem 2s-Timeout.
func PostgresHealth(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("pool not initialized")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var one int
	if err := pool.QueryRow(ctx, "SELECT 1").Scan(&one); err != nil {
		return fmt.Errorf("select 1: %w", err)
	}
	if one != 1 {
		return fmt.Errorf("unexpected result: %d", one)
	}
	return nil
}
