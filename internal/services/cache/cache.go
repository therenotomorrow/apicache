package cache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/therenotomorrow/apicache/internal/domain"
	"github.com/therenotomorrow/apicache/pkg/drivers"
)

const (
	pingWindow     = 10
	defaultMaxConn = 1
	defaultTimeout = time.Millisecond
)

var (
	ErrInvalidMaxConn     = errors.New("invalid MaxConn")
	ErrInvalidConnTimeout = errors.New("invalid ConnTimeout")
)

type (
	Driver interface {
		Get(ctx context.Context, key string) (string, error)
		Set(ctx context.Context, key string, val string) error
		Del(ctx context.Context, key string) error
		io.Closer
	}
	Config struct {
		MaxConn     int
		ConnTimeout time.Duration
	}
	Cache struct {
		driver Driver
		cfg    Config
		once   sync.Once
		keys   sync.Map
		done   chan struct{}
		queue  chan struct{}
	}
)

func New(cfg Config, driver Driver) (*Cache, error) {
	if cfg.MaxConn < defaultMaxConn {
		return nil, ErrInvalidMaxConn
	}

	if cfg.ConnTimeout < defaultTimeout {
		return nil, ErrInvalidConnTimeout
	}

	return &Cache{
		driver: driver,
		cfg:    cfg,
		once:   sync.Once{},
		keys:   sync.Map{},
		done:   make(chan struct{}),
		queue:  make(chan struct{}, cfg.MaxConn),
	}, nil
}

func MustNew(cfg Config, driver Driver) *Cache {
	obj, err := New(cfg, driver)
	if err != nil {
		panic(err)
	}

	return obj
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	err := c.acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer c.release()

	now := time.Now().UTC()

	deadline, ok := c.keys.Load(key)
	if !ok {
		return nil, domain.ErrKeyNotExist
	}

	// don't allow read expired keys, GC will remove it
	future, _ := deadline.(time.Time)
	if !future.IsZero() && now.After(future) {
		return nil, domain.ErrKeyExpired
	}

	// we assume that external driver also will not contain key because of `followEx()`
	raw, err := c.driver.Get(ctx, key)
	if errors.Is(err, drivers.ErrNotExist) {
		return nil, domain.ErrKeyNotExist
	}

	if err != nil {
		return nil, fmt.Errorf("driver error: %w", err)
	}

	return []byte(raw), nil
}

func (c *Cache) Set(ctx context.Context, key string, val []byte, deadline time.Time) error {
	err := c.acquire(ctx)
	if err != nil {
		return err
	}
	defer c.release()

	err = c.driver.Set(ctx, key, string(val))
	if err != nil {
		return fmt.Errorf("driver error: %w", err)
	}

	// set infinite key
	previous, _ := c.keys.Swap(key, deadline)

	if deadline.IsZero() {
		return nil
	}

	// if last key was zero - we need to run GC on it
	last, _ := previous.(time.Time)
	if last.IsZero() {
		ttl := deadline.Sub(time.Now().UTC())

		go c.followEx(context.WithoutCancel(ctx), key, ttl/pingWindow)
	}

	return nil
}

func (c *Cache) Del(ctx context.Context, key string) error {
	err := c.acquire(ctx)
	if err != nil {
		return err
	}
	defer c.release()

	err = c.driver.Del(ctx, key)
	if err != nil {
		return fmt.Errorf("driver error: %w", err)
	}

	c.keys.Delete(key)

	return nil
}

func (c *Cache) Close() error {
	var err error

	c.once.Do(func() {
		close(c.done)
		close(c.queue)

		err = c.driver.Close()
	})

	if err != nil {
		return fmt.Errorf("driver error: %w", err)
	}

	return nil
}

func (c *Cache) acquire(ctx context.Context) error {
	select {
	case <-c.done:
		return domain.ErrClosed
	default:
	}

	timer := time.NewTimer(c.cfg.ConnTimeout)
	defer timer.Stop()

	select {
	case c.queue <- struct{}{}:
		return nil
	case <-timer.C:
		return domain.ErrConnTimeout
	case <-ctx.Done():
		return domain.ErrContextTimeout
	}
}

func (c *Cache) release() {
	<-c.queue
}

func (c *Cache) followEx(ctx context.Context, key string, ping time.Duration) {
	ticker := time.NewTicker(ping)
	defer ticker.Stop()

	for tick := range ticker.C {
		deadline, ok := c.keys.Load(key)
		if !ok {
			return
		}

		future, _ := deadline.(time.Time)
		// no reason to wait for not-ex keys
		if future.IsZero() {
			return
		}

		if future.After(tick) {
			continue
		}

		// we will not stop GC if driver cause error
		err := c.driver.Del(ctx, key)
		if err != nil {
			continue
		}

		c.keys.Delete(key)

		break
	}
}
