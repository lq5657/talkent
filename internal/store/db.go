package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	if _, err := db.Exec(Schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize schema: %w", err)
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`ALTER TABLE analysis_reports ADD COLUMN markdown_content TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE analysis_reports ADD COLUMN model_used TEXT NOT NULL DEFAULT ''`,
	}
	for _, m := range migrations {
		_, _ = db.Exec(m)
	}
	return nil
}
