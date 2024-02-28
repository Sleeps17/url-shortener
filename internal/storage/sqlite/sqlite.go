package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"url-shortener/internal/cache"
	mapCache "url-shortener/internal/cache/map-cache"
	"url-shortener/internal/storage"
)

type Store struct {
	db    *sql.DB
	cache cache.Cache
}

func MustNew(ctx context.Context, storagePath string, capacity int) *Store {

	c, err := mapCache.New(capacity)
	if err != nil {
		panic(err)
	}

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

	return &Store{db: db, cache: c}
}

func (s *Store) Close(ctx context.Context) error {

	res := make(chan error)
	defer close(res)

	go func() {
		err1 := s.cache.Close(ctx)
		err2 := s.db.Close()

		if err1 != nil && err2 != nil {
			res <- fmt.Errorf("%w && %w", err1, err2)
		} else if err1 != nil {
			res <- err1
		} else if err2 != nil {
			res <- err2
		}

		res <- nil
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("time is out")
	case err := <-res:
		return err
	}
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

	res, err := s.db.ExecContext(ctx, query, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if cnt, _ := res.RowsAffected(); cnt == int64(0) {
		return storage.ErrAliasNotFound
	}

	return nil
}

func (s *Store) UpdateAlias(ctx context.Context, alias, newAlias string) error {
	const op = "sqlite.UpdateAlias"

	query := `UPDATE urls SET alias = ? WHERE alias = ?`

	res, err := s.db.ExecContext(ctx, query, newAlias, alias)
	if err != nil {
		var sqlErr sqlite3.Error
		if errors.As(err, &sqlErr) && errors.Is(sqlErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return storage.ErrNewAliasAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if cnt, _ := res.RowsAffected(); cnt == int64(0) {
		return storage.ErrAliasNotFound
	}

	return nil
}
