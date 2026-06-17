package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, db *pgxpool.Pool, dir string) error {
	return runSQLFiles(ctx, db, dir, "schema_migrations")
}

func RunSeeds(ctx context.Context, db *pgxpool.Pool, dir string) error {
	return runSQLFiles(ctx, db, dir, "seed_migrations")
}

func runSQLFiles(ctx context.Context, db *pgxpool.Pool, dir string, table string) error {
	if _, err := db.Exec(ctx, fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			name TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`, table)); err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		var exists bool
		if err := db.QueryRow(ctx, fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %s WHERE name = $1)", table), name).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}

		sqlBytes, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}

		tx, err := db.Begin(ctx)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("%s failed: %w", name, err)
		}

		if _, err := tx.Exec(ctx, fmt.Sprintf("INSERT INTO %s (name) VALUES ($1)", table), name); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}

	return nil
}
