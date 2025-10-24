package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/Parapheen/ph-clone/internal/infra/sqlite"
	"github.com/Parapheen/ph-clone/internal/pkg/config"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		logger.Warn("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize database
	db, err := sqlite.InitDB(cfg.Database.URL)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Check migration status
	logger.Info("Checking migration status...")

	// Check if schema_migrations table exists
	var tableExists bool
	err = db.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM sqlite_master 
		WHERE type='table' AND name='schema_migrations'
	`).Scan(&tableExists)

	if err != nil {
		logger.Error("Failed to check migration table", "error", err)
		os.Exit(1)
	}

	if !tableExists {
		logger.Info("No migration tracking found - this appears to be a fresh database")
		logger.Info("All migrations will be applied on next startup")
	} else {
		// Get applied migrations
		rows, err := db.Query("SELECT version, applied_at FROM schema_migrations ORDER BY version")
		if err != nil {
			logger.Error("Failed to query applied migrations", "error", err)
			os.Exit(1)
		}
		defer rows.Close()

		logger.Info("Applied migrations:")
		count := 0
		for rows.Next() {
			var version, appliedAt string
			if err := rows.Scan(&version, &appliedAt); err != nil {
				logger.Error("Failed to scan migration row", "error", err)
				continue
			}
			logger.Info("Migration applied", "version", version, "applied_at", appliedAt)
			count++
		}

		if count == 0 {
			logger.Info("No migrations have been applied yet")
		} else {
			logger.Info("Total migrations applied", "count", count)
		}
	}

	// Check if main tables exist
	tables := []string{"users", "products", "launches", "categories"}
	logger.Info("Checking main tables...")

	for _, table := range tables {
		var exists bool
		err = db.QueryRow(`
			SELECT COUNT(*) > 0 
			FROM sqlite_master 
			WHERE type='table' AND name=?
		`, table).Scan(&exists)

		if err != nil {
			logger.Error("Failed to check table", "table", table, "error", err)
			continue
		}

		if exists {
			logger.Info("Table exists", "table", table)
		} else {
			logger.Warn("Table missing", "table", table)
		}
	}
}
