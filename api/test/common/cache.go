package common

import (
	"context"
	"sync"
	"time"

	"atom-maintenance/internal/domain"
)

type MemCache struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemCache() *MemCache {
	return &MemCache{data: make(map[string]string)}
}

func (c *MemCache) Set(_ context.Context, key, value string, _ time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	return nil
}

func (c *MemCache) Get(_ context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.data[key]; ok {
		return v, nil
	}
	return "", domain.ErrCacheMiss
}

func (c *MemCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}
