package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Options func(*pgxpool.Config)

func WithMaxConns(n int32) Options {
	return func(cfg *pgxpool.Config) {
		cfg.MaxConns = n
	}
}

func WithMinConns(n int32) Options {
	return func(cfg *pgxpool.Config) {
		cfg.MinConns = n
	}
}

func WithMaxConnLifetime(t time.Duration) Options {
	return func(cfg *pgxpool.Config) {
		cfg.MaxConnLifetime = t
	}
}

func WithMaxConnIdleTime(t time.Duration) Options {
	return func(cfg *pgxpool.Config) {
		cfg.MaxConnIdleTime = t
	}
}
