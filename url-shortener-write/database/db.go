package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func Migrate(db *sql.DB, files fs.FS) error {
	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		body, err := fs.ReadFile(files, entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		for _, stmt := range splitStatements(string(body)) {
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

func splitStatements(script string) []string {
	parts := strings.Split(script, ";")
	stmts := make([]string, 0, len(parts))
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt != "" {
			stmts = append(stmts, stmt)
		}
	}
	return stmts
}
