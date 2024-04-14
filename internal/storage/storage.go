package storage

import (
	"context"
	"errors"
)

// Storage interface for storage
type Storage interface {
	// SaveURL saves {url} by {alias}
	SaveURL(ctx context.Context, url, alias, username string) error
	// GetURL returns {url} by {alias}
	GetURL(ctx context.Context, username, alias string) (string, error)
	// DeleteURL deletes {url} by {alias}
	DeleteURL(ctx context.Context, username, alias string) error
	//UpdateAlias replaces {alias} for {url}
	UpdateAlias(ctx context.Context, username, oldAlias, newAlias string) error

	Close(ctx context.Context) error
}

var (
	ErrAliasNotFound         = errors.New("alias not found")
	ErrAliasAlreadyExist     = errors.New("alias already exist")
	ErrNewAliasAlreadyExists = errors.New("new_alias cannot use, url with this alias already exists")
)

var (
	ErrCacheSet    = errors.New("failed to save url in cache")
	ErrCacheGet    = errors.New("failed to get url in cache")
	ErrCacheUpdate = errors.New("failed to update cache")
	ErrCacheDelete = errors.New("failed to delete alias from cache")
)
