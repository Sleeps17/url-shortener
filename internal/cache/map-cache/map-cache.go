package mapCache

import (
	"context"
	"fmt"
	"math"
	"sync"
	"url-shortener/internal/cache"
)

type Cache struct {
	store     map[cache.KeyType]cache.ValueType
	capacity  int
	frequency map[cache.KeyType]int
	mu        sync.RWMutex
}

func MustNew(capacity int) *Cache {
	return &Cache{
		store:     make(map[cache.KeyType]cache.ValueType),
		capacity:  capacity,
		frequency: make(map[cache.KeyType]int),
	}
}

func (c *Cache) Close(ctx context.Context) error {
	return nil
}

func (c *Cache) Set(ctx context.Context, url, alias, username string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	done := make(chan struct{})
	defer close(done)

	key := cache.KeyType{Username: username, Alias: alias}

	go func() {
		if _, ok := c.store[key]; ok {
			c.frequency[key]++
			c.store[key] = cache.ValueType{Url: url}

			done <- struct{}{}
			return
		}

		if c.capacity <= len(c.store) {
			var leastUsageKey cache.KeyType
			minUsage := math.MinInt

			for key, usage := range c.frequency {
				if usage < minUsage {
					leastUsageKey = key
					minUsage = usage
				}
			}

			delete(c.store, leastUsageKey)
			delete(c.frequency, leastUsageKey)
		}

		c.store[key] = cache.ValueType{Url: url}
		c.frequency[key]++

		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return cache.ErrTimeExceeded
	case <-done:
		return nil
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
		if value, ok := c.store[key]; ok {
			c.frequency[key]++

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

	done := make(chan struct{})
	defer close(done)

	key := cache.KeyType{Username: username, Alias: oldAlias}

	go func() {
		if value, ok := c.store[key]; ok {
			delete(c.store, key)
			newKey := cache.KeyType{Username: username, Alias: newAlias}
			c.store[newKey] = value

			done <- struct{}{}
			return
		}
	}()

	select {
	case <-ctx.Done():
		return cache.ErrTimeExceeded
	case <-done:
		return nil
	}
}

func (c *Cache) Delete(ctx context.Context, username, alias string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	done := make(chan struct{})

	key := cache.KeyType{Username: username, Alias: alias}

	go func() {
		delete(c.store, key)
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return cache.ErrTimeExceeded
	case <-done:
		return nil
	}

}
