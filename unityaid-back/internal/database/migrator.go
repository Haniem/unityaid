package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

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

		statements, err := splitSQLStatements(string(sqlBytes))
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("%s parse failed: %w", name, err)
		}

		for index, statement := range statements {
			if _, err := tx.Exec(ctx, statement); err != nil {
				_ = tx.Rollback(ctx)
				return fmt.Errorf("%s statement %d failed: %w", name, index+1, err)
			}
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

func splitSQLStatements(script string) ([]string, error) {
	var statements []string
	var current strings.Builder
	var dollarTag string
	inSingleQuote := false
	inDoubleQuote := false
	inLineComment := false
	inBlockComment := false

	for i := 0; i < len(script); i++ {
		ch := script[i]
		var next byte
		if i+1 < len(script) {
			next = script[i+1]
		}

		if inLineComment {
			current.WriteByte(ch)
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}

		if inBlockComment {
			current.WriteByte(ch)
			if ch == '*' && next == '/' {
				current.WriteByte(next)
				i++
				inBlockComment = false
			}
			continue
		}

		if dollarTag != "" {
			current.WriteByte(ch)
			if strings.HasPrefix(script[i:], dollarTag) {
				for j := 1; j < len(dollarTag); j++ {
					current.WriteByte(script[i+j])
				}
				i += len(dollarTag) - 1
				dollarTag = ""
			}
			continue
		}

		if inSingleQuote {
			current.WriteByte(ch)
			if ch == '\'' {
				if next == '\'' {
					current.WriteByte(next)
					i++
				} else {
					inSingleQuote = false
				}
			}
			continue
		}

		if inDoubleQuote {
			current.WriteByte(ch)
			if ch == '"' {
				if next == '"' {
					current.WriteByte(next)
					i++
				} else {
					inDoubleQuote = false
				}
			}
			continue
		}

		switch {
		case ch == '-' && next == '-':
			current.WriteByte(ch)
			current.WriteByte(next)
			i++
			inLineComment = true
		case ch == '/' && next == '*':
			current.WriteByte(ch)
			current.WriteByte(next)
			i++
			inBlockComment = true
		case ch == '\'':
			current.WriteByte(ch)
			inSingleQuote = true
		case ch == '"':
			current.WriteByte(ch)
			inDoubleQuote = true
		case ch == '$':
			if tag, ok := readDollarQuoteTag(script[i:]); ok {
				current.WriteString(tag)
				i += len(tag) - 1
				dollarTag = tag
			} else {
				current.WriteByte(ch)
			}
		case ch == ';':
			statement := strings.TrimSpace(current.String())
			if statement != "" {
				statements = append(statements, statement)
			}
			current.Reset()
		default:
			current.WriteByte(ch)
		}
	}

	if inSingleQuote || inDoubleQuote || inBlockComment || dollarTag != "" {
		return nil, fmt.Errorf("unterminated SQL literal or comment")
	}

	statement := strings.TrimSpace(current.String())
	if statement != "" {
		statements = append(statements, statement)
	}

	return statements, nil
}

func readDollarQuoteTag(input string) (string, bool) {
	if len(input) < 2 || input[0] != '$' {
		return "", false
	}
	for i := 1; i < len(input); i++ {
		if input[i] == '$' {
			return input[:i+1], true
		}
		if i == 1 && unicode.IsDigit(rune(input[i])) {
			return "", false
		}
		if input[i] != '_' && !unicode.IsLetter(rune(input[i])) && !unicode.IsDigit(rune(input[i])) {
			return "", false
		}
	}
	return "", false
}
