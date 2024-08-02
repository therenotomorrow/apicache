package memcached_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/therenotomorrow/apicache/internal/services/cache"
	"github.com/therenotomorrow/apicache/pkg/drivers"
	"github.com/therenotomorrow/apicache/pkg/drivers/memcached"
	"github.com/therenotomorrow/apicache/test/toolkit"
)

const (
	addr        = "127.0.0.1:11211"
	invalidAddr = ":0"
)

var (
	errDriverGet = errors.New("Memcached.Get() error: dial tcp :0: connect: connection refused")
	errDriverSet = errors.New("Memcached.Set() error: dial tcp :0: connect: connection refused")
	errDriverDel = errors.New("Memcached.Del() error: dial tcp :0: connect: connection refused")
)

func config() memcached.Config {
	return memcached.Config{Addr: addr}
}

func TestMain(m *testing.M) {
	client := memcache.New(addr)

	_ = client.FlushAll()

	_ = client.Set(&memcache.Item{Key: "insertKey", Value: []byte("insertVal"), Flags: 0, Expiration: 0, CasID: 0})
	_ = client.Set(&memcache.Item{Key: "updateKey", Value: []byte("updateVal"), Flags: 0, Expiration: 0, CasID: 0})
	_ = client.Set(&memcache.Item{Key: "deleteKey", Value: []byte("deleteVal"), Flags: 0, Expiration: 0, CasID: 0})

	m.Run()
}

func TestUnitNewWithConfig(t *testing.T) {
	t.Parallel()

	var _ cache.Driver = memcached.NewWithConfig(config())
}

func TestIntegrationMemcachedGet(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg memcached.Config
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
			args: args{cfg: memcached.Config{Addr: invalidAddr}, key: "key"},
			want: toolkit.Want("", errDriverGet),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			obj := memcached.NewWithConfig(test.args.cfg)

			got, err := obj.Get(ctx, test.args.key)

			toolkit.Assert(t, toolkit.Got(err, got), test.want)
		})
	}
}

func TestIntegrationMemcachedSet(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg memcached.Config
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
			args: args{cfg: memcached.Config{Addr: invalidAddr}, key: "key", val: ""},
			want: toolkit.Err(errDriverSet),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			obj := memcached.NewWithConfig(test.args.cfg)

			err := obj.Set(ctx, test.args.key, test.args.val)

			toolkit.Assert(t, toolkit.Got[any](err), test.want)

			if test.name == "driver error" {
				return
			}

			// check inner client data
			client := memcache.New(addr)
			item, _ := client.Get(test.args.key)

			assert.Equal(t, test.args.val, string(item.Value))
			assert.Empty(t, item.Expiration)
		})
	}
}

func TestIntegrationMemcachedDel(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg memcached.Config
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
			args: args{cfg: memcached.Config{Addr: invalidAddr}, key: "key"},
			want: toolkit.Err(errDriverDel),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			obj := memcached.NewWithConfig(test.args.cfg)

			err := obj.Del(ctx, test.args.key)

			toolkit.Assert(t, toolkit.Got[any](err), test.want)

			if test.name == "driver error" {
				return
			}

			// check inner client data
			client := memcache.New(addr)
			item, _ := client.Get(test.args.key)

			assert.Empty(t, item)
		})
	}
}

func TestUnitMemcachedClose(t *testing.T) {
	t.Parallel()

	obj := memcached.NewWithConfig(config())

	require.NoError(t, obj.Close())
}
