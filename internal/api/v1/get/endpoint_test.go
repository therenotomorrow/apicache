package apiv1get_test

import (
	"context"
	"errors"
	"github.com/kxnes/go-interviews/apicache/pkg/cache"
	"reflect"
	"testing"
)

const (
	Smoke1 = "smoke1"
	Smoke2 = "smoke2"
	Smoke3 = "smoke3"
	Smoke4 = "smoke4"
	Smoke5 = "smoke5"
	Smoke6 = "smoke6"
	Smoke7 = "smoke7"
	Smoke8 = "smoke8"
	Smoke9 = "smoke9"
)

var errDummy = errors.New("dummy error")

type cacheGetter struct{}

func (c cacheGetter) Get(_ context.Context, key string) (cache.ValType, error) {
	switch key {
	case Smoke2:
		return nil, cache.ErrKeyExpired
	case Smoke3:
		return nil, cache.ErrKeyNotExist
	case Smoke4:
		return nil, cache.ErrConnTimeout
	case Smoke5:
		return nil, cache.ErrContextTimeout
	case Smoke6:
		return nil, errDummy
	default:
		return map[string]any{"hello": "world", "age": 42}, nil
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	type params struct {
		names  []string
		values []string
	}

	type args struct {
		params *params
	}

	type want struct {
		code int
		body string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Get(tt.args.cacher); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
