package redis

import (
	"context"
	"errors"
	"sync"

	"github.com/kxnes/go-interviews/apicache/pkg/cache"
	"github.com/redis/go-redis/v9"
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
		return "", cache.ErrKeyNotExist
	}

	if err != nil {
		return "", cache.DriverError(err)
	}

	return val, nil
}

func (d *Redis) Set(ctx context.Context, key string, val string) error {
	_, err := d.client.Set(ctx, key, val, 0).Result()

	return cache.DriverError(err)
}

func (d *Redis) Del(ctx context.Context, key string) error {
	_, err := d.client.Del(ctx, key).Result()

	return cache.DriverError(err)
}

func (d *Redis) Close() error {
	var err error

	d.once.Do(func() {
		err = d.client.Close()
	})

	return cache.DriverError(err)
}

func (d *Redis) UnsafeClient() *redis.Client {
	return d.client
}
