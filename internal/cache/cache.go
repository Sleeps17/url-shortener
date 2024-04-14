package cache

import (
	"context"
	"encoding/json"
	"errors"
)

type Cache interface {
	Set(ctx context.Context, url, alias, username string) error
	Get(ctx context.Context, username, alias string) (string, error)
	Update(ctx context.Context, username, oldAlias, newAlias string) error
	Delete(ctx context.Context, username, alias string) error
	Close(ctx context.Context) error
}

type KeyType struct {
	Username string `json:"username" redis:"username"`
	Alias    string `json:"alias" redis:"alias"`
}

type ValueType struct {
	Url string `json:"url" redis:"url"`
}

func (v ValueType) MarshalBinary() ([]byte, error) {
	return json.Marshal(v)
}

var (
	ErrTimeExceeded = errors.New("time is out")
)
