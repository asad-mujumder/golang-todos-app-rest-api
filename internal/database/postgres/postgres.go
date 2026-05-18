package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConns          = int32(20)
	defaultMinConns          = int32(2)
	defaultMaxConnLifetime   = time.Hour
	defaultMaxConnIdleTime   = 30 * time.Minute
	defaultHealthCheckPeriod = time.Minute
	defaultConnectTimeout    = 5 * time.Second
)

func NewPool(databaseURL string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)

	if err != nil {
		return nil, fmt.Errorf("postgres: parse config: %w", err)
	}

	poolConfig.MaxConns        = defaultMaxConns
	poolConfig.MinConns        = defaultMinConns
	poolConfig.MaxConnLifetime = defaultMaxConnLifetime
	poolConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	poolConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)

	if err != nil {
		return nil, fmt.Errorf("postgres: create pool: %w", err)
	}

	if err := ping(pool); err != nil {
		pool.Close()
		return nil, err
	}
	
	return pool, nil
}

func ping(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultConnectTimeout)

	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("postgres: ping: %w", err)
	}

	return nil
}