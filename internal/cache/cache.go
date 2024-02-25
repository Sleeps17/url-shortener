package cache

import (
	"context"
	"errors"
)

type Cache interface {
	Set(ctx context.Context, url, alias string) error
	Get(ctx context.Context, alias string) (string, error)
	Update(ctx context.Context, oldAlias, newAlias string) error
	Delete(ctx context.Context, alias string) error
	Close(ctx context.Context) error
}

var (
	ErrTimeExceeded = errors.New("time is out")
)
