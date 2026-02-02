package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(connString string) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) GetPool() *pgxpool.Pool { 
	return p.pool
}

// GetDB returns a *sql.DB for compatibility with code expecting database/sql
func (p *Postgres) GetDB() *sql.DB {
	// Register and open using pgx stdlib
	db := stdlib.OpenDB(*p.pool.Config().ConnConfig)
	return db
}

func (p *Postgres) Close() {
	p.pool.Close()
}

// Exec executes a SQL statement (for migrations)
func (p *Postgres) Exec(ctx context.Context, sql string) error {
	_, err := p.pool.Exec(ctx, sql)
	return err
}
