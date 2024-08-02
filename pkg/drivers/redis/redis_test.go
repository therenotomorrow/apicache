package redis_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kxnes/go-interviews/apicache/internal/services/cache"
	"github.com/kxnes/go-interviews/apicache/pkg/drivers"
	"github.com/kxnes/go-interviews/apicache/pkg/drivers/redis"
	"github.com/kxnes/go-interviews/apicache/test/toolkit"
	redislib "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	addr        = "127.0.0.1:6379"
	invalidAddr = ":0"
)

var (
	errDriverGet    = errors.New("Redis.Get() error: dial tcp :0: connect: connection refused")
	errDriverSet    = errors.New("Redis.Set() error: dial tcp :0: connect: connection refused")
	errDriverDel    = errors.New("Redis.Del() error: dial tcp :0: connect: connection refused")
	errDriverClosed = fmt.Errorf("Redis.Close() error: %w", redislib.ErrClosed)
)

func config() redis.Config {
	return redis.Config{Addr: addr}
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	options := new(redislib.Options)
	options.Addr = addr

	client := redislib.NewClient(options)

	_ = client.FlushAll(ctx)

	_ = client.Set(ctx, "insertKey", "insertVal", 0)
	_ = client.Set(ctx, "updateKey", "updateVal", 0)
	_ = client.Set(ctx, "deleteKey", "deleteVal", 0)

	m.Run()
}

func TestUnitNewWithConfig(t *testing.T) {
	t.Parallel()

	var _ cache.Driver = redis.NewWithConfig(config())
}

func TestIntegrationRedisGet(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg redis.Config
		key string
	}

	tests := []struct {
		name string
		args args
		want toolkit.W[string]
	}{
		{
			name: "success",
			args: args{cfg: config(), key: "insertKey"},
			want: toolkit.Want("insertVal", nil),
		},
		{
			name: "key not exist",
			args: args{cfg: config(), key: "invalidKey"},
			want: toolkit.Want("", drivers.ErrNotExist),
		},
		{
			name: "driver error",
			args: args{cfg: redis.Config{Addr: invalidAddr}, key: "key"},
			want: toolkit.Want("", errDriverGet),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			obj := redis.NewWithConfig(test.args.cfg)

			got, err := obj.Get(ctx, test.args.key)

			toolkit.Assert(t, toolkit.Got(err, got), test.want)
		})
	}
}

func TestIntegrationRedisSet(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg redis.Config
		key string
		val string
	}

	tests := []struct {
		name string
		args args
		want toolkit.W[any]
	}{
		{
			name: "success",
			args: args{cfg: config(), key: "newKey", val: "newVal"},
			want: toolkit.Err(nil),
		},
		{
			name: "key exist",
			args: args{cfg: config(), key: "updateKey", val: "newVal"},
			want: toolkit.Err(nil),
		},
		{
			name: "driver error",
			args: args{cfg: redis.Config{Addr: invalidAddr}, key: "key", val: ""},
			want: toolkit.Err(errDriverSet),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			obj := redis.NewWithConfig(test.args.cfg)

			err := obj.Set(ctx, test.args.key, test.args.val)

			toolkit.Assert(t, toolkit.Got[any](err), test.want)

			if test.name == "driver error" {
				return
			}

			// check inner client data
			client := redislib.NewClient(new(redislib.Options))
			val := client.Get(ctx, test.args.key).Val()
			ttl := client.TTL(ctx, test.args.key).Val()

			assert.Equal(t, test.args.val, val)
			assert.Equal(t, -time.Nanosecond, ttl)
		})
	}
}

func TestIntegrationRedisDel(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg redis.Config
		key string
	}

	tests := []struct {
		name string
		args args
		want toolkit.W[any]
	}{
		{
			name: "success",
			args: args{cfg: config(), key: "deleteKey"},
			want: toolkit.Err(nil),
		},
		{
			name: "key not exist",
			args: args{cfg: config(), key: "invalidKey"},
			want: toolkit.Err(nil),
		},
		{
			name: "driver error",
			args: args{cfg: redis.Config{Addr: invalidAddr}, key: "key"},
			want: toolkit.Err(errDriverDel),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			obj := redis.NewWithConfig(test.args.cfg)

			err := obj.Del(ctx, test.args.key)

			toolkit.Assert(t, toolkit.Got[any](err), test.want)

			if test.name == "driver error" {
				return
			}

			// check inner client data
			client := redislib.NewClient(new(redislib.Options))
			val := client.Get(ctx, test.args.key).Val()

			assert.Empty(t, val)
		})
	}
}

func TestUnitRedisClose(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg redis.Config
	}

	tests := []struct {
		name string
		args args
		want toolkit.W[any]
	}{
		{name: "success", args: args{cfg: config()}, want: toolkit.Err(nil)},
		{name: "failure", args: args{cfg: config()}, want: toolkit.Err(errDriverClosed)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			obj := redis.NewWithConfig(test.args.cfg)

			if test.name == "failure" {
				_ = obj.UnsafeClient().Close()
			}

			err := obj.Close()

			toolkit.Assert(t, toolkit.Got[any](err), test.want)

			// calling Close() twice is safely do nothing
			require.NoError(t, obj.Close())
		})
	}
}
