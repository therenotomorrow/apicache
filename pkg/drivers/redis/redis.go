package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/therenotomorrow/apicache/pkg/drivers"
)

type (
	Config struct {
		Addr string
	}
	Redis struct {
		cfg    Config
		once   sync.Once
		client *redis.Client
	}
)

func NewWithConfig(cfg Config) *Redis {
	options := new(redis.Options)
	options.Addr = cfg.Addr

	return &Redis{cfg: cfg, once: sync.Once{}, client: redis.NewClient(options)}
}

func (d *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := d.client.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return "", drivers.ErrNotExist
	}

	if err != nil {
		return "", fmt.Errorf("Redis.Get() error: %w", err)
	}

	return val, nil
}

func (d *Redis) Set(ctx context.Context, key string, val string) error {
	_, err := d.client.Set(ctx, key, val, 0).Result()
	if err != nil {
		return fmt.Errorf("Redis.Set() error: %w", err)
	}

	return nil
}

func (d *Redis) Del(ctx context.Context, key string) error {
	_, err := d.client.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("Redis.Del() error: %w", err)
	}

	return nil
}

func (d *Redis) Close() error {
	var err error

	d.once.Do(func() {
		err = d.client.Close()
	})

	if err != nil {
		return fmt.Errorf("Redis.Close() error: %w", err)
	}

	return nil
}

func (d *Redis) UnsafeClient() *redis.Client {
	return d.client
}
