// Package db owns the pgx connection pool lifecycle and a small transaction
// helper. Domain modules use pgx directly — there is no generic repository
// framework.
package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skycode/ojt-management/backend/internal/config"
)

// Pool wraps pgxpool with the project's configuration.
type Pool struct {
	*pgxpool.Pool
}

func NewPool(ctx context.Context, cfg config.DBConfig) (*Pool, error) {
	pc, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}
	pc.MaxConns = cfg.MaxConns
	pc.MinConns = cfg.MinConns
	pc.MaxConnLifetime = cfg.MaxConnLifetime
	pc.MaxConnIdleTime = cfg.MaxConnIdleTime
	pc.ConnConfig.ConnectTimeout = 10 * time.Second

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &Pool{pool}, nil
}

// Health pings the database with a bounded timeout.
func (p *Pool) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return p.Ping(ctx)
}

func (p *Pool) Shutdown() { p.Close() }

// InTx runs fn inside a transaction, committing on success.
func (p *Pool) InTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// IsUniqueViolation reports whether err is a PostgreSQL unique-constraint
// violation, optionally restricted to a named constraint.
func IsUniqueViolation(err error, constraint string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return false
	}
	return constraint == "" || pgErr.ConstraintName == constraint
}
