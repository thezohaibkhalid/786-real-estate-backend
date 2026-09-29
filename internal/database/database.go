package database

import (
	"context"
	"example.com/786-real-estate/backend/internal/config"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func Open(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	pc, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid database configuration")
	}
	pc.MaxConns = cfg.MaxConns
	pc.ConnConfig.ConnectTimeout = 5 * time.Second
	pc.MaxConnLifetime = time.Hour
	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		return nil, fmt.Errorf("database pool initialization failed")
	}
	check, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(check); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database unavailable; check DATABASE_URL and PostgreSQL")
	}
	return pool, nil
}
