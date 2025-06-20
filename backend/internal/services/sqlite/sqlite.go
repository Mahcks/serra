package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/mahcks/serra/internal/db/repository"
	_ "github.com/mattn/go-sqlite3"
)

type SetupOptions struct {
	Path string // Path to SQLite file, e.g., "./data.db"
}

func Setup(ctx context.Context, Version string, opts SetupOptions) (Service, error) {
	// Check if file exists; if not, Go/SQLite will create it on connect
	if _, err := os.Stat(opts.Path); os.IsNotExist(err) {
		fmt.Println("SQLite database does not exist; will be created on first connect.")
	}

	db, err := sql.Open("sqlite3", opts.Path+"?_fk=1&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	svc := &sqliteService{
		db:      db,
		queries: repository.New(db),
	}

	return svc, nil
}
