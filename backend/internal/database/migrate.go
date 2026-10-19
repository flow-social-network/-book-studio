package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	dir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR"))
	if dir == "" { dir = "migrations" }
	if err := pool.QueryRow(ctx, `SELECT 1`).Scan(new(int)); err != nil { return fmt.Errorf("database ping before migration: %w", err) }
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil { return fmt.Errorf("create migration ledger: %w", err) }
	const version = "0001_initial"
	var applied bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, version).Scan(&applied); err != nil { return fmt.Errorf("read migration ledger: %w", err) }
	if applied { return nil }
	sql, err := os.ReadFile(filepath.Join(dir, "0001_initial.sql"))
	if err != nil { return fmt.Errorf("read migration %s: %w", version, err) }
	tx, err := pool.Begin(ctx)
	if err != nil { return fmt.Errorf("begin migration: %w", err) }
	defer func(){ _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, string(sql)); err != nil { return fmt.Errorf("apply migration %s: %w", version, err) }
	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, version); err != nil { return fmt.Errorf("record migration %s: %w", version, err) }
	if err := tx.Commit(ctx); err != nil { return fmt.Errorf("commit migration %s: %w", version, err) }
	return nil
}
