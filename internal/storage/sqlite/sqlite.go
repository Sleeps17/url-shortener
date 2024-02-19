package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"url-shortener/internal/storage"
)

type Store struct {
	db *sql.DB
}

func MustNew(ctx context.Context, storagePath string) *Store {

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	query := `CREATE TABLE IF NOT EXISTS "urls" (
		"id" INTEGER PRIMARY KEY,
		"alias" TEXT NOT NULL UNIQUE,
		"url" TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias on urls(alias);	
`
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		panic(err)
	}

	if _, err := stmt.Exec(); err != nil {
		panic(err)
	}

	return &Store{db: db}
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) SaveURL(ctx context.Context, url, alias string) error {
	const op = "sqlite.SaveURL"

	query := `INSERT INTO urls (alias, url) VALUES (?, ?);`

	_, err := s.db.ExecContext(ctx, query, alias, url)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return storage.ErrAliasAlreadyExist
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Store) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "sqlite.GetURL"

	query := `SELECT url FROM urls WHERE alias = ?`

	rows, err := s.db.QueryContext(ctx, query, alias)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	defer func() { _ = rows.Close() }()

	var url string
	for rows.Next() {
		if err := rows.Scan(&url); err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}
	}

	if url == "" {
		return "", storage.ErrAliasNotFound
	}

	return url, nil
}

func (s *Store) DeleteURL(ctx context.Context, alias string) error {
	const op = "sqlite.DeleteURL"

	query := `DELETE FROM urls WHERE alias = ?`

	_, err := s.db.ExecContext(ctx, query, alias)
	if err != nil {
		if errors.Is(err, sqlite3.ErrNotFound) {
			return storage.ErrAliasNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Store) UpdateAlias(ctx context.Context, url, newAlias string) error {
	const op = "sqlite.UpdateAlias"

	query := `UPDATE urls SET alias = ? WHERE url = ?`

	_, err := s.db.ExecContext(ctx, query, newAlias, url)
	if err != nil {
		if errors.Is(err, sqlite3.ErrNotFound) {
			return storage.ErrUrlNotFound
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
