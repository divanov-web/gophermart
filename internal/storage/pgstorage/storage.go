package pgstorage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	return pool, nil
}

func NewStorage(ctx context.Context, pool *pgxpool.Pool) (*Storage, error) {
	pgStorage := &Storage{
		pool: pool,
	}

	return pgStorage, nil
}

func (s *Storage) Ping() error {
	return s.pool.Ping(context.Background())
}
