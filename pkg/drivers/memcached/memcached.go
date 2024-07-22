package memcached

import (
	"context"
	"errors"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/kxnes/go-interviews/apicache/pkg/cache"
)

type (
	Config struct {
		Addr string
	}
	Memcached struct {
		cfg    Config
		client *memcache.Client
	}
)

func NewWithConfig(cfg Config) *Memcached {
	return &Memcached{cfg: cfg, client: memcache.New(cfg.Addr)}
}

func (d *Memcached) Get(_ context.Context, key string) (string, error) {
	item, err := d.client.Get(key)

	if errors.Is(err, memcache.ErrCacheMiss) {
		return "", cache.ErrKeyNotExist
	}

	if err != nil {
		return "", cache.DriverError(err)
	}

	return string(item.Value), nil
}

func (d *Memcached) Set(_ context.Context, key string, val string) error {
	item := &memcache.Item{
		Key:        key,
		Value:      []byte(val),
		Flags:      0,
		Expiration: 0,
	}

	err := d.client.Set(item)

	return cache.DriverError(err)
}

func (d *Memcached) Del(_ context.Context, key string) error {
	err := d.client.Delete(key)

	if errors.Is(err, memcache.ErrCacheMiss) {
		return nil
	}

	return cache.DriverError(err)
}

func (d *Memcached) Close() error {
	return nil
}
