package postgres

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	io.Closer
	driver.Pinger

	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)

	Begin(ctx context.Context, f func(tx pgx.Tx) error) error
}

func New(ctx context.Context, dsn string, opts ...Options) (DB, error) {
	cfg, err := parseConfig(dsn, opts...)
	if err != nil {
		return nil, fmt.Errorf("parse psql config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create psql pool: %v", err)
	}

	client := &db{db: pool}

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping psql: %v", err)
	}

	return client, nil
}

func parseConfig(dsn string, opts ...Options) (*pgxpool.Config, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	completeConfig(cfg)

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg, nil
}

func completeConfig(cfg *pgxpool.Config) {
	if cfg.MaxConns == 0 {
		cfg.MaxConns = 10
	}
	if cfg.MinConns == 0 {
		cfg.MinConns = 1
	}
	if cfg.HealthCheckPeriod == 0 {
		cfg.HealthCheckPeriod = 5
	}
}
