package cache

import (
	"context"
	"time"
)

// Store is the minimal cache interface used by application services.
// *Cache (Redis) and standalone.MemCache both satisfy this interface.
type Store interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string, dest any) error
	Delete(ctx context.Context, keys ...string) error
}
