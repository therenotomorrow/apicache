package machine_test

import (
	"context"
	"testing"

	"github.com/kxnes/go-interviews/apicache/pkg/cache"
	"github.com/kxnes/go-interviews/apicache/pkg/drivers/machine"
	"github.com/kxnes/go-interviews/apicache/test/toolkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var instance = machine.New() //nolint:gochecknoglobals

func TestUnitNew(t *testing.T) {
	t.Parallel()

	var _ cache.Driver = machine.New()
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	_ = instance.Set(ctx, "insertKey", "insertVal")
	_ = instance.Set(ctx, "updateKey", "updateVal")
	_ = instance.Set(ctx, "deleteKey", "deleteVal")

	m.Run()
}

func TestUnitMachineGet(t *testing.T) {
	t.Parallel()

	type args struct {
		key string
	}

	tests := []struct {
		name string
		args args
		want toolkit.W[string]
	}{
		{name: "success", args: args{key: "insertKey"}, want: toolkit.Want("insertVal", nil)},
		{name: "key not exist", args: args{key: "invalidKey"}, want: toolkit.Want("", cache.ErrKeyNotExist)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			got, err := instance.Get(ctx, test.args.key)

			toolkit.Assert(t, toolkit.Got(err, got), test.want)
		})
	}
}

func TestUnitMachineSet(t *testing.T) {
	t.Parallel()

	type args struct {
		key string
		val string
	}

	tests := []struct {
		name string
		args args
		want toolkit.W[any]
	}{
		{name: "success", args: args{key: "newKey", val: "newVal"}, want: toolkit.Err(nil)},
		{name: "key exist", args: args{key: "updateKey", val: "newVal"}, want: toolkit.Err(nil)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			err := instance.Set(ctx, test.args.key, test.args.val)

			toolkit.Assert(t, toolkit.Got[any](err), test.want)

			// check inner data
			val, _ := instance.Get(ctx, test.args.key)

			assert.Equal(t, test.args.val, val)
		})
	}
}

func TestUnitMachineDel(t *testing.T) {
	t.Parallel()

	type args struct {
		key string
	}

	tests := []struct {
		name string
		args args
		want toolkit.W[any]
	}{
		{name: "success", args: args{key: "deleteKey"}, want: toolkit.Err(nil)},
		{name: "key not exist", args: args{key: "invalidKey"}, want: toolkit.Err(nil)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			err := instance.Del(ctx, test.args.key)

			toolkit.Assert(t, toolkit.Got[any](err), test.want)

			// check inner data
			val, _ := instance.Get(ctx, test.args.key)

			assert.Empty(t, val)
		})
	}
}

func TestUnitMachineClose(t *testing.T) {
	t.Parallel()

	obj := machine.New()

	require.NoError(t, obj.Close())
}
