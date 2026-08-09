package migration

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"regexp"
	"sort"

	"github.com/jackc/pgx/v5"
)

// upMigrations embeds every "*.up.sql" file in this directory so the
// compiled binary is self-contained: the Docker image needs nothing beyond
// the binary itself to bring a fresh database up to date.
//
//go:embed *.up.sql
var upMigrations embed.FS

// versionPattern extracts the leading numeric version from a migration
// filename, e.g. "001_blog_table.up.sql" -> "001".
var versionPattern = regexp.MustCompile(`^(\d+)_.*\.up\.sql$`)

// migrationLockID is an arbitrary, fixed key for pg_advisory_lock: any two
// processes calling Run concurrently (e.g. two instances starting at once)
// serialize on this lock instead of racing to apply the same migration
// twice.
const migrationLockID = 8743028 // arbitrary constant, unique enough for this app

// Run applies any pending "up" migrations embedded in this package, in
// ascending version order, tracking applied versions in a schema_migrations
// table so calling Run again (e.g. on every app startup) is a no-op once a
// migration has been applied. Each migration's SQL and its bookkeeping
// insert run in a single transaction, so a failure partway through doesn't
// leave a migration recorded as applied when it wasn't.
//
// Down migrations (*.down.sql) are intentionally not run here — they exist
// for manual/ops rollback (e.g. `psql -f ...`), not automatic startup.
func Run(ctx context.Context, conn *pgx.Conn) error {
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationLockID); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		if _, err := conn.Exec(ctx, "SELECT pg_advisory_unlock($1)", migrationLockID); err != nil {
			slog.Error("failed to release migration lock", "error", err)
		}
	}()

	if err := ensureSchemaMigrationsTable(ctx, conn); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	applied, err := appliedVersions(ctx, conn)
	if err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}

	pending, err := pendingMigrations(applied)
	if err != nil {
		return err
	}

	for _, m := range pending {
		if err := applyMigration(ctx, conn, m); err != nil {
			return err
		}
		slog.Info("applied migration", "version", m.version, "file", m.filename)
	}

	return nil
}

type migrationFile struct {
	version  string
	filename string
}

// pendingMigrations lists embedded migration files not present in applied,
// sorted by ascending numeric version.
func pendingMigrations(applied map[string]bool) ([]migrationFile, error) {
	entries, err := upMigrations.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	var pending []migrationFile
	for _, entry := range entries {
		name := entry.Name()
		matches := versionPattern.FindStringSubmatch(name)
		if matches == nil {
			continue
		}

		version := matches[1]
		if applied[version] {
			continue
		}

		pending = append(pending, migrationFile{version: version, filename: name})
	}

	sort.Slice(pending, func(i, j int) bool { return pending[i].version < pending[j].version })
	return pending, nil
}

func applyMigration(ctx context.Context, conn *pgx.Conn, m migrationFile) error {
	sqlBytes, err := upMigrations.ReadFile(m.filename)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", m.filename, err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction for migration %s: %w", m.filename, err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op once Commit succeeds

	if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("apply migration %s: %w", m.filename, err)
	}

	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, m.version); err != nil {
		return fmt.Errorf("record migration %s: %w", m.filename, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", m.filename, err)
	}

	return nil
}

func ensureSchemaMigrationsTable(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version     TEXT PRIMARY KEY,
		applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	return err
}

func appliedVersions(ctx context.Context, conn *pgx.Conn) (map[string]bool, error) {
	rows, err := conn.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}
