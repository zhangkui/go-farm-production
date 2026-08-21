// Package migrate applies versioned SQL migration files embedded in the
// binary. Each file is executed as a single transaction; the schema_migrations
// table records which versions have been applied.
package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed sql/*.sql
var sqlFS embed.FS

// Migration is one versioned SQL file.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// Load reads and orders migrations from the embedded sql directory.
func Load() ([]Migration, error) {
	entries, err := fs.ReadDir(sqlFS, "sql")
	if err != nil {
		return nil, fmt.Errorf("migrate: read embed dir: %w", err)
	}
	var ms []Migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		parts := strings.SplitN(e.Name(), "_", 2)
		if len(parts) < 2 {
			continue
		}
		ver, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("migrate: parse version %q: %w", e.Name(), err)
		}
		data, err := sqlFS.ReadFile("sql/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("migrate: read %s: %w", e.Name(), err)
		}
		ms = append(ms, Migration{Version: ver, Name: e.Name(), SQL: string(data)})
	}
	sort.Slice(ms, func(i, j int) bool { return ms[i].Version < ms[j].Version })
	return ms, nil
}

// Runner applies migrations against a database.
type Runner struct {
	db *sql.DB
}

// NewRunner creates a migration runner.
func NewRunner(db *sql.DB) *Runner { return &Runner{db: db} }

// EnsureSchemaMigrations creates the bookkeeping table if absent.
func (r *Runner) EnsureSchemaMigrations(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    BIGINT   NOT NULL,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (version)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	return err
}

// Applied returns the set of already-applied versions.
func (r *Runner) Applied(ctx context.Context) (map[int]bool, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out[v] = true
	}
	return out, rows.Err()
}

// Up applies every pending migration in order. Each migration runs in its own
// transaction; on failure the transaction rolls back and Up returns an error.
func (r *Runner) Up(ctx context.Context) ([]Migration, error) {
	if err := r.EnsureSchemaMigrations(ctx); err != nil {
		return nil, fmt.Errorf("migrate: ensure schema_migrations: %w", err)
	}
	applied, err := r.Applied(ctx)
	if err != nil {
		return nil, fmt.Errorf("migrate: read applied: %w", err)
	}
	migrations, err := Load()
	if err != nil {
		return nil, err
	}
	var ran []Migration
	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}
		if err := r.apply(ctx, m); err != nil {
			return ran, fmt.Errorf("migrate: apply %s: %w", m.Name, err)
		}
		ran = append(ran, m)
	}
	return ran, nil
}

func (r *Runner) apply(ctx context.Context, m Migration) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Split the script into individual statements and execute each one. The
	// go-sql-driver/mysql driver disallows multi-statement queries by default
	// (multiStatements off), so a whole-file Exec fails with a syntax error on
	// the second statement. Splitting here keeps that safety default in place.
	for _, stmt := range splitStatements(m.SQL) {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("exec: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (version) VALUES (?) ON DUPLICATE KEY UPDATE version = version`,
		m.Version); err != nil {
		return fmt.Errorf("record version: %w", err)
	}
	return tx.Commit()
}

// splitStatements breaks a SQL script into individual statements, respecting
// single/double-quoted string literals, backtick identifiers and line (`--`,
// `#`) and block (`/* */`) comments so that terminators inside them are not
// mistaken for statement boundaries. Byte-level iteration is safe for UTF-8
// because continuation/lead bytes (>=0x80) never collide with ASCII
// punctuation. Each returned statement is trimmed of surrounding whitespace;
// comments-only fragments are dropped.
func splitStatements(script string) []string {
	var (
		out []string
		buf strings.Builder
		i   = 0
		n   = len(script)
	)
	flush := func() {
		s := strings.TrimSpace(buf.String())
		if s != "" {
			out = append(out, s)
		}
		buf.Reset()
	}
	for i < n {
		c := script[i]
		// line comment: -- or # ... to end of line
		if (c == '-' && i+1 < n && script[i+1] == '-') || c == '#' {
			for i < n && script[i] != '\n' {
				i++
			}
			continue
		}
		// block comment: /* ... */
		if c == '/' && i+1 < n && script[i+1] == '*' {
			i += 2
			for i+1 < n && !(script[i] == '*' && script[i+1] == '/') {
				i++
			}
			i += 2
			continue
		}
		// single-quoted string literal
		if c == '\'' {
			buf.WriteByte(c)
			i++
			for i < n {
				if script[i] == '\\' && i+1 < n { // escaped char
					buf.WriteByte(script[i])
					buf.WriteByte(script[i+1])
					i += 2
					continue
				}
				if script[i] == '\'' {
					buf.WriteByte('\'')
					i++
					if i < n && script[i] == '\'' { // '' = literal quote
						buf.WriteByte('\'')
						i++
						continue
					}
					break
				}
				buf.WriteByte(script[i])
				i++
			}
			continue
		}
		// backtick identifier
		if c == '`' {
			buf.WriteByte(c)
			i++
			for i < n {
				if script[i] == '`' {
					buf.WriteByte('`')
					i++
					if i < n && script[i] == '`' { // `` = literal backtick
						buf.WriteByte('`')
						i++
						continue
					}
					break
				}
				buf.WriteByte(script[i])
				i++
			}
			continue
		}
		// double-quoted (identifier or string, mode-dependent)
		if c == '"' {
			buf.WriteByte(c)
			i++
			for i < n {
				if script[i] == '\\' && i+1 < n {
					buf.WriteByte(script[i])
					buf.WriteByte(script[i+1])
					i += 2
					continue
				}
				if script[i] == '"' {
					buf.WriteByte('"')
					i++
					if i < n && script[i] == '"' {
						buf.WriteByte('"')
						i++
						continue
					}
					break
				}
				buf.WriteByte(script[i])
				i++
			}
			continue
		}
		// statement terminator
		if c == ';' {
			flush()
			i++
			continue
		}
		buf.WriteByte(c)
		i++
	}
	flush() // trailing statement without terminator
	return out
}
