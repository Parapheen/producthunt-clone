package sqlite

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dsn string) (*sqlx.DB, error) {
	if !strings.Contains(dsn, "_journal_mode") {
		dsn += "?_journal_mode=WAL"
	} else { // Already has params
		dsn += "&_journal_mode=WAL"
	}
	if !strings.Contains(dsn, "_foreign_keys") {
		dsn += "&_foreign_keys=on"
	}
	if !strings.Contains(dsn, "_busy_timeout") {
		dsn += "&_busy_timeout=5000" // 5 seconds
	}

	db, err := sqlx.Connect("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("could not init db driver for dsn %q: %w", dsn, err)
	}

	// --- Connection Pool Tuning ---
	// IMPORTANT for SQLite: Limit open connections to 1 to prevent "database is locked"
	// errors during concurrent writes, even with WAL mode. Read operations can still
	// happen concurrently with WAL. Increase only if you *know* you won't have
	// concurrent writes or have implemented external locking.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)                // Usually matches MaxOpenConns
	db.SetConnMaxLifetime(1 * time.Hour) // Example: Recycle connections periodically

	// --- Verify Connection ---
	// Ping also ensures PRAGMAs applied via DSN are valid.
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("could not connect to db %q: %w", dsn, err)
	}

	return db, nil
}
