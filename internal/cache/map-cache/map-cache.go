package mapCache

import (
	"context"
	"fmt"
	"math"
	"sync"
	"url-shortener/internal/cache"
)

type Cache struct {
	store     map[string]string
	capacity  int
	frequency map[string]int
	mu        sync.RWMutex
}

func New(capacity int) (*Cache, error) {
	return &Cache{
		store:     make(map[string]string),
		capacity:  capacity,
		frequency: make(map[string]int),
	}, nil
}

func (c *Cache) Close(ctx context.Context) error {
	return nil
}

func (c *Cache) Set(ctx context.Context, url, alias string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	done := make(chan struct{})
	defer close(done)

	go func() {
		if _, ok := c.store[alias]; ok {
			c.frequency[alias]++
			c.store[alias] = url

			done <- struct{}{}
			return
		}

		if c.capacity <= len(c.store) {
			leastUsageAlias := ""
			minUsage := math.MinInt

			for alias, usage := range c.frequency {
				if usage < minUsage {
					leastUsageAlias = alias
					minUsage = usage
				}
			}

			delete(c.store, leastUsageAlias)
			delete(c.frequency, leastUsageAlias)
		}

		c.store[alias] = url
		c.frequency[alias]++

		done <- struct{}{}
		return
	}()

	select {
	case <-ctx.Done():
		return cache.ErrTimeExceeded
	case <-done:
		return nil
	}
}

func (c *Cache) Get(ctx context.Context, alias string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	res := make(chan struct {
		string
		error
	})
	defer close(res)

	go func() {
		if url, ok := c.store[alias]; ok {
			c.frequency[alias]++

			res <- struct {
				string
				error
			}{string: url, error: nil}
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

func (c *Cache) Update(ctx context.Context, oldAlias, newAlias string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	done := make(chan struct{})
	defer close(done)

	go func() {
		if url, ok := c.store[oldAlias]; ok {
			delete(c.store, oldAlias)
			c.store[newAlias] = url

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

func (c *Cache) Delete(ctx context.Context, alias string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	done := make(chan struct{})

	go func() {
		delete(c.store, alias)
		done <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return cache.ErrTimeExceeded
	case <-done:
		return nil
	}

}
