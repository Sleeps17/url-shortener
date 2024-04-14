package redisCache

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"math"
	"sync"
	"time"
	"url-shortener/internal/cache"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	client    *redis.Client
	capacity  int
	frequency map[cache.KeyType]int
	mu        sync.RWMutex
}

func MustNew(addr string, db int, timeout time.Duration, capacity int) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		panic(err)
	}

	return &Cache{
		client:    client,
		capacity:  capacity,
		frequency: make(map[cache.KeyType]int),
	}
}

func (c *Cache) Close(ctx context.Context) error {
	return c.client.Close()
}

func (c *Cache) Set(ctx context.Context, url, alias, username string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	done := make(chan error)
	defer close(done)

	key := cache.KeyType{Username: username, Alias: alias}

	go func() {
		keyData, err := EncodeKey(key)
		if err != nil {
			done <- err
			return
		}
		if _, err := c.client.Get(ctx, keyData).Result(); err == nil {
			c.frequency[key]++
			if err := c.client.Set(ctx, keyData, cache.ValueType{Url: url}, 0); err != nil {
				done <- err.Err()
				return
			}

			done <- nil
			return
		}

		size, err := c.client.DBSize(ctx).Result()
		if err != nil {
			done <- err
			return
		}
		if int64(c.capacity) <= size {
			var leastUsageKey cache.KeyType
			minUsage := math.MinInt

			for key, usage := range c.frequency {
				if usage < minUsage {
					leastUsageKey = key
					minUsage = usage
				}
			}

			leastUsageKeyData, err := EncodeKey(leastUsageKey)
			if err != nil {
				done <- err
				return
			}

			if _, err := c.client.Del(ctx, leastUsageKeyData).Result(); err != nil {
				done <- err
				return
			}
			delete(c.frequency, leastUsageKey)
		}

		if err := c.client.Set(ctx, keyData, cache.ValueType{Url: url}, 0); err != nil {
			done <- err.Err()
			return
		}
		c.frequency[key]++

		done <- nil
	}()

	select {
	case <-ctx.Done():
		return cache.ErrTimeExceeded
	case err := <-done:
		return err
	}
}

func (c *Cache) Get(ctx context.Context, username, alias string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	res := make(chan struct {
		string
		error
	})
	defer close(res)

	key := cache.KeyType{Username: username, Alias: alias}

	go func() {
		keyData, err := EncodeKey(key)
		if err != nil {
			res <- struct {
				string
				error
			}{"", err}
			return
		}

		if valueData, err := c.client.Get(ctx, keyData).Bytes(); err == nil {
			c.frequency[key]++

			value, err := DecodeValue(string(valueData))
			if err != nil {
				res <- struct {
					string
					error
				}{"", err}
				return
			}

			res <- struct {
				string
				error
			}{string: value.Url, error: nil}
			return
		}

		res <- struct {
			string
			error
		}{string: "", error: fmt.Errorf("alias not found")}
	}()

	select {
	case <-ctx.Done():
		return "", cache.ErrTimeExceeded
	case r := <-res:
		return r.string, r.error
	}
}

func (c *Cache) Update(ctx context.Context, username, oldAlias, newAlias string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	done := make(chan error)
	defer close(done)

	oldKey := cache.KeyType{Username: username, Alias: oldAlias}
	newKey := cache.KeyType{Username: username, Alias: newAlias}

	go func() {
		oldKeyData, err := EncodeKey(oldKey)
		if err != nil {
			done <- err
			return
		}

		newKeyData, err := EncodeKey(newKey)
		if err != nil {
			done <- err
			return
		}

		if _, err := c.client.Rename(ctx, oldKeyData, newKeyData).Result(); err != nil {
			done <- err
			return
		}
		done <- nil
	}()

	select {
	case <-ctx.Done():
		return cache.ErrTimeExceeded
	case err := <-done:
		return err
	}
}

func (c *Cache) Delete(ctx context.Context, username, alias string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	done := make(chan error)
	defer close(done)

	key := cache.KeyType{Username: username, Alias: alias}

	go func() {

		keyData, err := EncodeKey(key)
		if err != nil {
			done <- err
			return
		}

		if _, err := c.client.Del(ctx, keyData).Result(); err != nil {
			done <- err
			return
		}

		done <- nil
	}()

	select {
	case <-ctx.Done():
		return cache.ErrTimeExceeded
	case err := <-done:
		return err
	}
}

func EncodeKey(key cache.KeyType) (string, error) {
	var buff bytes.Buffer

	err := gob.NewEncoder(&buff).Encode(key)

	if err != nil {
		return "", err
	}

	return buff.String(), nil
}

func EncodeValue(value cache.ValueType) (string, error) {
	var buff bytes.Buffer

	err := gob.NewEncoder(&buff).Encode(value)

	if err != nil {
		return "", err
	}

	return buff.String(), nil
}

func DecodeKey(data string) (cache.KeyType, error) {

	var key cache.KeyType
	err := gob.NewDecoder(bytes.NewReader([]byte(data))).Decode(&key)
	if err != nil {
		return cache.KeyType{}, err
	}

	return key, err
}

func DecodeValue(data string) (cache.ValueType, error) {

	var value cache.ValueType
	err := gob.NewDecoder(bytes.NewReader([]byte(data))).Decode(&value)
	if err != nil {
		return cache.ValueType{}, err
	}

	return value, err
}
