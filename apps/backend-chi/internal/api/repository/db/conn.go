package db

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func migrationsDirURL() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("cannot resolve migrations path")
	}

	path := filepath.Join(filepath.Dir(file), "query", "migrations")
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return "file://" + filepath.ToSlash(abs), nil
}

func NewPool(connStr string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func ClosePool(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}

func MigrateDB(connStr string) error {
	migrationsURL, err := migrationsDirURL()
	if err != nil {
		return err
	}

	m, err := migrate.New(migrationsURL, connStr)
	if err != nil {
		return err
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
