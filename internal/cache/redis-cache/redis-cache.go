package redisCache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"math"
	"sync"
	"time"
	"url-shortener/internal/cache"
)

type Cache struct {
	client    *redis.Client
	capacity  int
	frequency map[string]int
	mu        sync.RWMutex
}

func New(addr string, db int, timeout time.Duration, capacity int) (*Cache, error) {
	const op = "redis-cache.New"
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Cache{
		client:    client,
		capacity:  capacity,
		frequency: make(map[string]int),
	}, nil
}

func (c *Cache) Close(ctx context.Context) error {
	return c.client.Close()
}

func (c *Cache) Set(ctx context.Context, url, alias string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	done := make(chan error)
	defer close(done)

	go func() {
		if _, err := c.client.Get(ctx, alias).Result(); err == nil {
			c.frequency[alias]++
			if err := c.client.Set(ctx, alias, url, 0); err != nil {
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
			leastUsageAlias := ""
			minUsage := math.MinInt

			for alias, usage := range c.frequency {
				if usage < minUsage {
					leastUsageAlias = alias
					minUsage = usage
				}
			}

			if _, err := c.client.Del(ctx, leastUsageAlias).Result(); err != nil {
				done <- err
				return
			}
			delete(c.frequency, leastUsageAlias)
		}

		if err := c.client.Set(ctx, alias, url, 0); err != nil {
			done <- err.Err()
			return
		}
		c.frequency[alias]++

		done <- nil
		return
	}()

	select {
	case <-ctx.Done():
		return cache.ErrTimeExceeded
	case err := <-done:
		return err
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
		if url, err := c.client.Get(ctx, alias).Result(); err == nil {
			fmt.Println("FROM CACHE_____________________________________________")
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

	done := make(chan error)
	defer close(done)

	go func() {
		if _, err := c.client.Rename(ctx, oldAlias, newAlias).Result(); err != nil {
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

func (c *Cache) Delete(ctx context.Context, alias string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	done := make(chan error)
	defer close(done)

	go func() {
		if _, err := c.client.Del(ctx, alias).Result(); err != nil {
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
