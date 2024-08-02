package cache_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kxnes/go-interviews/apicache/internal/domain"
	"github.com/kxnes/go-interviews/apicache/internal/services/cache"
	"github.com/kxnes/go-interviews/apicache/pkg/drivers"
	"github.com/kxnes/go-interviews/apicache/test/mocks"
	"github.com/kxnes/go-interviews/apicache/test/toolkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	maxConn            = 1
	invalidMaxConn     = -1
	connTimeout        = 10 * time.Millisecond
	invalidConnTimeout = 0
)

var (
	errDummy        = errors.New("dummy error")
	errClosedDriver = errors.New("closed")
	errClosed       = fmt.Errorf("driver error: %w", errClosedDriver)
	errDummyDriver  = fmt.Errorf("driver error: %w", errDummy)
)

func value() []byte {
	return []byte(`{"age":42,"hello":"world"}`)
}

func config() cache.Config {
	return cache.Config{MaxConn: maxConn, ConnTimeout: connTimeout}
}

func driver() *mocks.DriverMock {
	return mocks.NewDriverMock()
}

func TestUnitNew(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg    cache.Config
		driver *mocks.DriverMock
	}

	tests := []struct {
		name string
		args args
		want toolkit.W[*cache.Cache]
	}{
		{
			name: "invalid MaxConn",
			args: args{cfg: cache.Config{MaxConn: invalidMaxConn, ConnTimeout: connTimeout}, driver: driver()},
			want: toolkit.Want[*cache.Cache](nil, cache.ErrInvalidMaxConn),
		},
		{
			name: "invalid ConnTimeout",
			args: args{cfg: cache.Config{MaxConn: maxConn, ConnTimeout: invalidConnTimeout}, driver: driver()},
			want: toolkit.Want[*cache.Cache](nil, cache.ErrInvalidConnTimeout),
		},
		{
			name: "success",
			args: args{cfg: config(), driver: driver()},
			want: toolkit.Want(new(cache.Cache), nil),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			obj, err := cache.New(test.args.cfg, test.args.driver)

			if test.name == "success" {
				assert.NotEmpty(t, obj)

				obj = new(cache.Cache)
			}

			toolkit.Assert(t, toolkit.Got(err, obj), test.want)
		})
	}
}

func TestUnitMustNew(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg    cache.Config
		driver *mocks.DriverMock
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{cfg: config(), driver: driver()},
		},
		{
			name: "failure",
			args: args{cfg: cache.Config{MaxConn: invalidMaxConn, ConnTimeout: invalidConnTimeout}, driver: driver()},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if test.name == "failure" {
				require.Panics(t, func() {
					_ = cache.MustNew(test.args.cfg, test.args.driver)
				})
			} else {
				require.NotPanics(t, func() {
					_ = cache.MustNew(test.args.cfg, test.args.driver)
				})
			}
		})
	}
}

func TestUnitCacheGetErrClosed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver())

	_ = obj.Close()

	got, err := obj.Get(ctx, "insertKey")

	toolkit.Assert(t, toolkit.Got(err, got), toolkit.Want[[]byte](nil, domain.ErrClosed))
}

func TestUnitCacheGetErrConnTimeout(t *testing.T) {
	t.Parallel()

	driver := driver()

	waiter := sync.WaitGroup{}
	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.SetMock = func(_ context.Context, _ string, _ string) error {
		time.Sleep(10 * connTimeout)

		return nil
	}
	driver.GetMock = func(_ context.Context, _ string) (string, error) {
		bytes, err := json.Marshal(map[string]any{"hello": "world", "age": 42})
		if err != nil {
			panic(err)
		}

		return string(bytes), nil
	}

	waiter.Add(1)

	go func() {
		_ = obj.Set(ctx, "insertKey", value(), time.Time{})

		waiter.Done()
	}()
	time.Sleep(connTimeout)

	got, err := obj.Get(ctx, "insertKey")

	toolkit.Assert(t, toolkit.Got(err, got), toolkit.Want[[]byte](nil, domain.ErrConnTimeout))

	waiter.Wait()

	got, err = obj.Get(ctx, "insertKey")

	toolkit.Assert(t, toolkit.Got(err, got), toolkit.Want(value(), nil))
}

func TestUnitCacheGetErrContextTimeout(t *testing.T) {
	t.Parallel()

	driver := driver()

	waiter := sync.WaitGroup{}
	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.SetMock = func(_ context.Context, _ string, _ string) error {
		time.Sleep(10 * connTimeout)

		return nil
	}
	driver.GetMock = func(_ context.Context, _ string) (string, error) {
		bytes, err := json.Marshal(map[string]any{"hello": "world", "age": 42})
		if err != nil {
			panic(err)
		}

		return string(bytes), nil
	}

	waiter.Add(1)

	go func() {
		_ = obj.Set(ctx, "insertKey", value(), time.Time{})

		waiter.Done()
	}()
	time.Sleep(connTimeout)

	newCtx, cancel := context.WithCancel(ctx)
	cancel()

	got, err := obj.Get(newCtx, "insertKey")

	toolkit.Assert(t, toolkit.Got(err, got), toolkit.Want[[]byte](nil, domain.ErrContextTimeout))

	waiter.Wait()

	got, err = obj.Get(ctx, "insertKey")

	toolkit.Assert(t, toolkit.Got(err, got), toolkit.Want(value(), nil))
}

func TestUnitCacheGetErrDriver(t *testing.T) {
	t.Parallel()

	driver := driver()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.GetMock = func(_ context.Context, _ string) (string, error) {
		return "", errDummy
	}

	_ = obj.Set(ctx, "insertKey", value(), time.Time{})

	got, err := obj.Get(ctx, "insertKey")

	toolkit.Assert(t, toolkit.Got(err, got), toolkit.Want[[]byte](nil, errDummyDriver))
}

func TestUnitCacheSetErrClosed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver())

	_ = obj.Close()

	err := obj.Set(ctx, "insertKey", value(), time.Time{})

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrClosed))
}

func TestUnitCacheSetErrConnTimeout(t *testing.T) {
	t.Parallel()

	driver := driver()

	waiter := sync.WaitGroup{}
	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.SetMock = func(_ context.Context, _ string, _ string) error {
		time.Sleep(10 * connTimeout)

		return nil
	}

	waiter.Add(1)

	go func() {
		_ = obj.Set(ctx, "insertKey", value(), time.Time{})

		waiter.Done()
	}()
	time.Sleep(connTimeout)

	err := obj.Set(ctx, "newInsertKey", value(), time.Time{})

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrConnTimeout))

	waiter.Wait()

	_, err = obj.Get(ctx, "newInsertKey")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrKeyNotExist))
}

func TestUnitCacheSetErrContextTimeout(t *testing.T) {
	t.Parallel()

	driver := driver()

	waiter := sync.WaitGroup{}
	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.SetMock = func(_ context.Context, _ string, _ string) error {
		time.Sleep(10 * connTimeout)

		return nil
	}

	waiter.Add(1)

	go func() {
		_ = obj.Set(ctx, "insertKey", value(), time.Time{})

		waiter.Done()
	}()
	time.Sleep(connTimeout)

	newCtx, cancel := context.WithCancel(ctx)
	cancel()

	err := obj.Set(newCtx, "newInsertKey", value(), time.Time{})

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrContextTimeout))

	waiter.Wait()

	_, err = obj.Get(ctx, "newInsertKey")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrKeyNotExist))
}

func TestUnitCacheSetErrDriver(t *testing.T) {
	t.Parallel()

	driver := driver()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.SetMock = func(_ context.Context, _ string, _ string) error {
		return errDummy
	}

	err := obj.Set(ctx, "insertKey", value(), time.Time{})

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(errDummyDriver))

	_, err = obj.Get(ctx, "insertKey")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrKeyNotExist))
}

func TestUnitCacheDelErrClosed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver())

	_ = obj.Close()

	err := obj.Del(ctx, "insertKey")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrClosed))
}

func TestUnitCacheDelErrConnTimeout(t *testing.T) {
	t.Parallel()

	driver := driver()

	waiter := sync.WaitGroup{}
	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.SetMock = func(_ context.Context, _ string, _ string) error {
		time.Sleep(10 * connTimeout)

		return nil
	}

	waiter.Add(1)

	go func() {
		_ = obj.Set(ctx, "insertKey", value(), time.Time{})

		waiter.Done()
	}()
	time.Sleep(connTimeout)

	err := obj.Del(ctx, "insertKey")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrConnTimeout))

	waiter.Wait()

	err = obj.Del(ctx, "insertKey")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(nil))
}

func TestUnitCacheDelErrContextTimeout(t *testing.T) {
	t.Parallel()

	driver := driver()

	waiter := sync.WaitGroup{}
	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.SetMock = func(_ context.Context, _ string, _ string) error {
		time.Sleep(10 * connTimeout)

		return nil
	}

	waiter.Add(1)

	go func() {
		_ = obj.Set(ctx, "insertKey", value(), time.Time{})

		waiter.Done()
	}()
	time.Sleep(connTimeout)

	newCtx, cancel := context.WithCancel(ctx)
	cancel()

	err := obj.Del(newCtx, "insertKey")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrContextTimeout))

	waiter.Wait()

	err = obj.Del(ctx, "insertKey")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(nil))
}

func TestUnitCacheDelErrDriver(t *testing.T) {
	t.Parallel()

	driver := driver()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.DelMock = func(_ context.Context, _ string) error {
		return errDummy
	}

	err := obj.Del(ctx, "insertKey")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(errDummyDriver))
}

func TestUnitCacheLogicNonExKey(t *testing.T) {
	t.Parallel()

	driver := driver()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.GetMock = func(_ context.Context, _ string) (string, error) {
		bytes, err := json.Marshal(map[string]any{"hello": "world", "age": 42})
		if err != nil {
			panic(err)
		}

		return string(bytes), nil
	}

	_ = obj.Set(ctx, "key", value(), time.Time{})

	got, err := obj.Get(ctx, "key")

	toolkit.Assert(t, toolkit.Got(err, got), toolkit.Want(value(), nil))
}

func TestUnitCacheLogicExKey(t *testing.T) {
	t.Parallel()

	driver := driver()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	_ = obj.Set(ctx, "key", value(), time.Now().UTC().Add(connTimeout))
	time.Sleep(2 * connTimeout)

	_, err := obj.Get(ctx, "key")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrKeyNotExist))
}

func TestUnitCacheLogicNonExKeyBecomesEx(t *testing.T) {
	t.Parallel()

	driver := driver()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.GetMock = func(_ context.Context, _ string) (string, error) {
		bytes, err := json.Marshal(map[string]any{"hello": "world", "age": 42})
		if err != nil {
			panic(err)
		}

		return string(bytes), nil
	}

	_ = obj.Set(ctx, "key", value(), time.Time{})
	_ = obj.Set(ctx, "key", value(), time.Now().UTC().Add(2*connTimeout))

	val, err := obj.Get(ctx, "key")

	toolkit.Assert(t, toolkit.Got(err, val), toolkit.Want(value(), nil))
	time.Sleep(4 * connTimeout)

	_, err = obj.Get(ctx, "key")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrKeyNotExist))
}

func TestUnitCacheLogicIncreaseExOnKey(t *testing.T) {
	t.Parallel()

	driver := driver()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.GetMock = func(_ context.Context, _ string) (string, error) {
		bytes, err := json.Marshal(map[string]any{"hello": "world", "age": 42})
		if err != nil {
			panic(err)
		}

		return string(bytes), nil
	}

	_ = obj.Set(ctx, "key", value(), time.Now().UTC().Add(5*connTimeout))
	time.Sleep(2 * connTimeout)

	_ = obj.Set(ctx, "key", value(), time.Now().UTC().Add(10*connTimeout))
	time.Sleep(6 * connTimeout)

	val, err := obj.Get(ctx, "key")

	toolkit.Assert(t, toolkit.Got(err, val), toolkit.Want(value(), nil))
	time.Sleep(6 * connTimeout)

	_, err = obj.Get(ctx, "key")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrKeyNotExist))
}

func TestUnitCacheLogicDecreaseExOnKey(t *testing.T) {
	t.Parallel()

	driver := driver()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.GetMock = func(_ context.Context, _ string) (string, error) {
		bytes, err := json.Marshal(map[string]any{"hello": "world", "age": 42})
		if err != nil {
			panic(err)
		}

		return string(bytes), nil
	}

	_ = obj.Set(ctx, "key", value(), time.Now().UTC().Add(10*connTimeout))
	time.Sleep(connTimeout)

	_ = obj.Set(ctx, "key", value(), time.Now().UTC().Add(5*connTimeout))
	time.Sleep(7 * connTimeout)

	_, err := obj.Get(ctx, "key")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrKeyNotExist))
}

func TestUnitCacheLogicExKeyBecomesNonEx(t *testing.T) {
	t.Parallel()

	driver := driver()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.GetMock = func(_ context.Context, _ string) (string, error) {
		bytes, err := json.Marshal(map[string]any{"hello": "world", "age": 42})
		if err != nil {
			panic(err)
		}

		return string(bytes), nil
	}

	_ = obj.Set(ctx, "key", value(), time.Now().UTC().Add(5*connTimeout))
	time.Sleep(2 * connTimeout)

	_ = obj.Set(ctx, "key", value(), time.Time{})

	time.Sleep(6 * connTimeout)

	val, err := obj.Get(ctx, "key")

	toolkit.Assert(t, toolkit.Got(err, val), toolkit.Want(value(), nil))
}

func TestUnitCacheLogicExKeyWasDeleted(t *testing.T) {
	t.Parallel()

	driver := driver()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	_ = obj.Set(ctx, "key", value(), time.Now().UTC().Add(5*connTimeout))
	time.Sleep(2 * connTimeout)

	_ = obj.Del(ctx, "key")

	time.Sleep(6 * connTimeout)

	_, err := obj.Get(ctx, "key")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrKeyNotExist))
}

func TestUnitCacheLogicExKeyWaitOnDeletion(t *testing.T) {
	t.Parallel()

	driver := driver()

	cnt := atomic.Int64{}
	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.GetMock = func(_ context.Context, _ string) (string, error) {
		bytes, err := json.Marshal(map[string]any{"hello": "world", "age": 42})
		if err != nil {
			panic(err)
		}

		return string(bytes), nil
	}
	driver.DelMock = func(_ context.Context, _ string) error {
		if cnt.Load() > 2 {
			return nil
		}

		cnt.Add(1)

		return errDummy
	}

	_ = obj.Set(ctx, "key", value(), time.Now().UTC().Add(2*connTimeout))
	time.Sleep(2 * connTimeout)

	_, err := obj.Get(ctx, "key")

	// key was expired but still not deleted
	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrKeyExpired))

	time.Sleep(4 * connTimeout)

	_, err = obj.Get(ctx, "key")

	// key was deleted successfully
	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrKeyNotExist))
}

func TestUnitCacheLogicKeyNotFoundInDriver(t *testing.T) {
	t.Parallel()

	driver := driver()

	ctx := context.Background()
	obj := cache.MustNew(config(), driver)

	driver.GetMock = func(_ context.Context, _ string) (string, error) {
		return "", drivers.ErrNotExist
	}

	_ = obj.Set(ctx, "key", value(), time.Time{})

	_, err := obj.Get(ctx, "key")

	toolkit.Assert(t, toolkit.Got[any](err), toolkit.Err(domain.ErrKeyNotExist))
}

func TestUnitCacheLogicSmoke(t *testing.T) {
	t.Parallel()

	driver := driver()

	waiter := sync.WaitGroup{}
	ctx := context.Background()
	obj := cache.MustNew(cache.Config{MaxConn: 20, ConnTimeout: connTimeout}, driver)

	driver.SetMock = func(_ context.Context, _ string, _ string) error {
		time.Sleep(connTimeout)

		return nil
	}
	driver.GetMock = func(_ context.Context, _ string) (string, error) {
		bytes, err := json.Marshal(map[string]any{"hello": "world", "age": 42})
		if err != nil {
			panic(err)
		}

		return string(bytes), nil
	}

	waiter.Add(100)

	success := atomic.Int64{}
	failure := atomic.Int64{}

	for idx := range 100 {
		go func(key string) {
			err := obj.Set(ctx, key, value(), time.Time{})
			if err != nil {
				failure.Add(1)
			} else {
				success.Add(1)
			}

			waiter.Done()
		}(strconv.Itoa(idx))
	}

	waiter.Wait()

	assert.Less(t, success.Load(), failure.Load())
	assert.LessOrEqual(t, success.Load(), int64(40))
	assert.GreaterOrEqual(t, failure.Load(), int64(60))
}

func TestUnitCacheClose(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg    cache.Config
		driver *mocks.DriverMock
	}

	tests := []struct {
		name string
		args args
		want toolkit.W[any]
	}{
		{name: "success", args: args{cfg: config(), driver: driver()}, want: toolkit.Err(nil)},
		{name: "failure", args: args{cfg: config(), driver: driver()}, want: toolkit.Err(errClosed)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			obj := cache.MustNew(test.args.cfg, test.args.driver)

			if test.name == "failure" {
				test.args.driver.CloseMock = func() error { return errClosedDriver }
			}

			err := obj.Close()

			toolkit.Assert(t, toolkit.Got[any](err), test.want)

			// calling Close() twice is safely do nothing
			require.NoError(t, obj.Close())
		})
	}
}
