package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"launchpad.icu/autopilot/database/sqlc"
)

type DB struct {
	*sqlc.Queries
	pool *pgxpool.Pool
}

func Open(ctx context.Context, uri string) (*DB, error) {
	pool, err := pgxpool.New(ctx, uri)
	if err != nil {
		return nil, err
	}

	return &DB{
		Queries: sqlc.New(pool),
		pool:    pool,
	}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) Migrate(path string) error {
	return goose.Up(stdlib.OpenDBFromPool(db.pool), path)
}
