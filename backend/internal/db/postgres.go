package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
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

func (p *Postgres) Close() {
	p.pool.Close()
}

// Exec executes a SQL statement (for migrations)
func (p *Postgres) Exec(ctx context.Context, sql string) error {
	_, err := p.pool.Exec(ctx, sql)
	return err
}
