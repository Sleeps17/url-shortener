package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"url-shortener/internal/cache"
	"url-shortener/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Store struct {
	db    *sql.DB
	cache cache.Cache
}

func MustNew(timeout time.Duration, c cache.Cache, storagePath string) *Store {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	defer cancel()

	mainFunc := func() *Store {

		db, err := sql.Open("sqlite3", storagePath)
		if err != nil {
			panic(err)
		}

		if err := db.Ping(); err != nil {
			panic(err)
		}

		query1 := `
			CREATE TABLE IF NOT EXISTS "users" (
				"id" INTEGER PRIMARY KEY,
				"username" TEXT NOT NULL UNIQUE
			);
		`

		query2 := `CREATE TABLE IF NOT EXISTS "urls" (
				"id" INTEGER PRIMARY KEY,
				"user_id" INT NOT NULL,
				"alias" TEXT NOT NULL,
				"url" TEXT NOT NULL,
				FOREIGN KEY (user_id) REFERENCES users(id),
				UNIQUE (user_id, alias)
			);`

		stmt, err := db.PrepareContext(ctx, query1)
		if err != nil {
			panic(err)
		}

		if _, err := stmt.Exec(); err != nil {
			panic(err)
		}

		stmt, err = db.PrepareContext(ctx, query2)
		if err != nil {
			panic(err)
		}

		if _, err := stmt.Exec(); err != nil {
			panic(err)
		}

		return &Store{db: db, cache: c}
	}

	return mainFunc()
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

func (s *Store) SaveURL(ctx context.Context, url, alias, username string) error {
	const op = "sqlite.SaveURL"

	var userId int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE username = ?", username).Scan(&userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			res, err := s.db.ExecContext(ctx, "INSERT INTO users (username) VALUES (?)", username)
			if err != nil {
				return fmt.Errorf("%s: failed to insert new user: %w", op, err)
			}

			userId, err = res.LastInsertId()
			if err != nil {
				return fmt.Errorf("%s: failed to get last inserted id: %w", op, err)
			}
		} else {
			return fmt.Errorf("%s: failed to select user_id: %w", op, err)
		}
	}

	query := `INSERT INTO urls (user_id, alias, url) VALUES (?, ?, ?);`

	_, err = s.db.ExecContext(ctx, query, userId, alias, url)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return fmt.Errorf("%s: %w", op, storage.ErrAliasAlreadyExist)
		}
		return fmt.Errorf("%s: failed to save (user_id, alias, url): %w", op, err)
	}

	if err := s.cache.Set(ctx, url, alias, username); err != nil {
		return fmt.Errorf("%s: %w: %w", op, storage.ErrCacheSet, err)
	}

	return nil
}

func (s *Store) GetURL(ctx context.Context, username, alias string) (string, error) {
	const op = "sqlite.GetURL"

	if url, err := s.cache.Get(ctx, username, alias); err == nil {
		return url, nil
	}

	query := `
		SELECT url
		FROM users AS u
		JOIN urls AS l ON u.id = l.user_id
		WHERE u.username = ? AND l.alias = ?
	`

	var url string
	err := s.db.QueryRowContext(ctx, query, username, alias).Scan(&url)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrAliasNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return url, nil
}

func (s *Store) DeleteURL(ctx context.Context, username, alias string) error {
	const op = "sqlite.DeleteURL"

	query := `DELETE FROM urls WHERE user_id = (SELECT id FROM users WHERE username = ?) AND alias = ?`

	res, err := s.db.ExecContext(ctx, query, username, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if cnt, _ := res.RowsAffected(); cnt == int64(0) {
		return storage.ErrAliasNotFound
	}

	if err := s.cache.Delete(ctx, username, alias); err != nil {
		return storage.ErrCacheDelete
	}

	return nil
}

func (s *Store) UpdateAlias(ctx context.Context, username, alias, newAlias string) error {
	const op = "sqlite.UpdateAlias"

	query := `UPDATE urls SET alias = ? WHERE user_id = (SELECT id FROM users WHERE username = ?) AND alias = ?`

	res, err := s.db.ExecContext(ctx, query, newAlias, username, alias)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return storage.ErrNewAliasAlreadyExists
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if cnt, _ := res.RowsAffected(); cnt == int64(0) {
		return storage.ErrAliasNotFound
	}

	if err := s.cache.Update(ctx, username, alias, newAlias); err != nil {
		return storage.ErrCacheUpdate
	}

	return nil
}
