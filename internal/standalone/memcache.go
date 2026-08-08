package standalone

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// MemCache is a simple in-memory cache that implements cache.Store.
// Used in standalone mode instead of Redis.
type MemCache struct {
	mu      sync.RWMutex
	entries map[string]memEntry
}

type memEntry struct {
	data      []byte
	expiresAt time.Time
}

func NewMemCache() *MemCache {
	mc := &MemCache{entries: make(map[string]memEntry)}
	go mc.cleanup()
	return mc
}

func (m *MemCache) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.entries[key] = memEntry{data: data, expiresAt: time.Now().Add(ttl)}
	m.mu.Unlock()
	return nil
}

func (m *MemCache) Get(_ context.Context, key string, dest any) error {
	m.mu.RLock()
	entry, ok := m.entries[key]
	m.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return errors.New("cache: chave não encontrada")
	}
	return json.Unmarshal(entry.data, dest)
}

func (m *MemCache) Delete(_ context.Context, keys ...string) error {
	m.mu.Lock()
	for _, k := range keys {
		delete(m.entries, k)
	}
	m.mu.Unlock()
	return nil
}

func (m *MemCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		m.mu.Lock()
		for k, e := range m.entries {
			if now.After(e.expiresAt) {
				delete(m.entries, k)
			}
		}
		m.mu.Unlock()
	}
}
