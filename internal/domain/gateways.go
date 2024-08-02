package domain

import (
	"context"
	"time"
)

type (
	CacheGetter interface {
		Get(ctx context.Context, key string) ([]byte, error)
	}
	CacheSetter interface {
		Set(ctx context.Context, key string, val []byte, deadline time.Time) error
	}
	CacheDeleter interface {
		Del(ctx context.Context, key string) error
	}
)
