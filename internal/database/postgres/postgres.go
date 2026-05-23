package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/asad-mujumder/golang-todos-app-rest-api/config"
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

func NewPool(ctx context.Context, dbConfig config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(dbConfig.DSN())

	if err != nil {
		return nil, fmt.Errorf("postgres: parse config: %w", err)
	}

	poolConfig.MaxConns        = defaultMaxConns
	poolConfig.MinConns        = defaultMinConns
	poolConfig.MaxConnLifetime = defaultMaxConnLifetime
	poolConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	poolConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)

	if err != nil {
		return nil, fmt.Errorf("postgres: create pool: %w", err)
	}

	if err := ping(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	
	return pool, nil
}

func ping(ctx context.Context, pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(ctx, defaultConnectTimeout)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("postgres: ping: %w", err)
	}

	return nil
}