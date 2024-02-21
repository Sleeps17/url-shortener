package storage

import (
	"context"
	"errors"
)

// Storage interface for storage
type Storage interface {
	// SaveURL saves {url} by {alias}
	SaveURL(ctx context.Context, url, alias string) error
	// GetURL returns {url} by {alias}
	GetURL(ctx context.Context, alias string) (string, error)
	// DeleteURL deletes {url} by {alias}
	DeleteURL(ctx context.Context, alias string) error
	//UpdateAlias replaces {alias} for {url}
	UpdateAlias(ctx context.Context, oldAlias, newAlias string) error
}

var (
	ErrAliasNotFound         = errors.New("alias not found")
	ErrAliasAlreadyExist     = errors.New("alias already exist")
	ErrNewAliasAlreadyExists = errors.New("new_alias cannot use, url with this alias already exists")
	ErrUrlNotFound           = errors.New("url not found")
)
