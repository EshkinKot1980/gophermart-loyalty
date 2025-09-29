package pg

import (
	"context"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func NewDB(ctx context.Context, dsn string) (*DB, error) {
	db := &DB{}

	m, err := migrate.New("file://db/migrations", dsn)
	if err != nil {
		return db, fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return db, fmt.Errorf("failed to apply migrations to the DB: %w", err)
	}

	db.pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		return db, fmt.Errorf("failed to create DB a connection pool: %w", err)
	}

	return db, nil
}

func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

func (db *DB) Close() {
	db.pool.Close()
}
