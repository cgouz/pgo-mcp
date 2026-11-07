package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, uri string, maxConns, minConns int32) (*pgxpool.Pool, error) {

	cfg, err := pgxpool.ParseConfig(uri)
	{
		if err != nil {
			return nil, fmt.Errorf("parse database config: %w", err)
		}

		cfg.MaxConns = maxConns
		cfg.MinConns = minConns

	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	{
		if err != nil {
			return nil, fmt.Errorf("create database pool: %w", err)
		}
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return pool, nil
}
