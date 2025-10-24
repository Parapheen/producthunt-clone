package sqlite

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migration represents a database migration
type Migration struct {
	Version string
	Up      string
	Down    string
}

// RunMigrations executes all pending migrations
func RunMigrations(db *sql.DB, migrationsDir string, logger *slog.Logger) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Read migration files
	migrations, err := readMigrations(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if _, applied := appliedMigrations[migration.Version]; !applied {
			logger.Info("Applying migration", "version", migration.Version)

			if err := applyMigration(db, migration); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
			}

			if err := recordMigration(db, migration.Version); err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
			}

			logger.Info("Migration applied successfully", "version", migration.Version)
		}
	}

	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := db.Exec(query)
	return err
}

func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
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

func readMigrations(migrationsDir string) ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir(migrationsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".sql") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		migration := parseMigration(string(content))
		if migration.Version == "" {
			// Extract version from filename (e.g., "20250615131409_state.sql" -> "20250615131409")
			filename := filepath.Base(path)
			parts := strings.Split(filename, "_")
			if len(parts) > 0 {
				migration.Version = parts[0]
			}
		}

		if migration.Version != "" {
			migrations = append(migrations, migration)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func parseMigration(content string) Migration {
	var migration Migration

	lines := strings.Split(content, "\n")
	var upLines []string
	var downLines []string
	var inUp, inDown bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "-- +goose Up") {
			inUp = true
			inDown = false
			continue
		}

		if strings.Contains(line, "-- +goose Down") {
			inUp = false
			inDown = true
			continue
		}

		if strings.Contains(line, "-- +goose StatementBegin") || strings.Contains(line, "-- +goose StatementEnd") {
			continue
		}

		if inUp && line != "" && !strings.HasPrefix(line, "--") {
			upLines = append(upLines, line)
		}

		if inDown && line != "" && !strings.HasPrefix(line, "--") {
			downLines = append(downLines, line)
		}
	}

	migration.Up = strings.Join(upLines, "\n")
	migration.Down = strings.Join(downLines, "\n")

	return migration
}

func applyMigration(db *sql.DB, migration Migration) error {
	if migration.Up == "" {
		return fmt.Errorf("no up migration found")
	}

	_, err := db.Exec(migration.Up)
	return err
}

func recordMigration(db *sql.DB, version string) error {
	query := "INSERT INTO schema_migrations (version) VALUES (?)"
	_, err := db.Exec(query, version)
	return err
}
