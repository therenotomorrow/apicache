package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kxnes/go-interviews/apicache/internal/domain"
	"github.com/kxnes/go-interviews/apicache/test/toolkit"
)

const (
	Smoke1 = "smoke1"
	Smoke2 = "smoke2"
	Smoke3 = "smoke3"
	Smoke4 = "smoke4"
	Smoke5 = "smoke5"
	Smoke6 = "smoke6"
	Smoke7 = "smoke7"
)

var errDummy = errors.New("dummy error")

type (
	getter        struct{}
	setter        struct{}
	deleter       struct{}
	cannotMarshal struct{}
)

func (c cannotMarshal) MarshalJSON() ([]byte, error) {
	return nil, errDummy
}

func (g getter) Get(_ context.Context, key string) ([]byte, error) {
	switch key {
	case Smoke3:
		return nil, errDummy
	case Smoke4:
		return []byte{}, nil
	}

	return []byte(`{"hello":"world","age":42}`), nil
}

func (s setter) Set(_ context.Context, key string, _ []byte, deadline time.Time) error {
	switch key {
	case Smoke1, Smoke3:
		if !deadline.IsZero() {
			return errDummy
		}

		return nil
	case Smoke2:
		if deadline.IsZero() {
			return errDummy
		}

		return nil
	case Smoke7:
		return errDummy
	}

	return nil
}

func (d deleter) Del(_ context.Context, key string) error {
	if key == Smoke3 {
		return errDummy
	}

	return nil
}

func TestUnitGetUseCase(t *testing.T) {
	t.Parallel()

	type args struct {
		key string
	}

	tests := []struct {
		name string
		args args
		want toolkit.W[domain.ValType]
	}{
		{
			name: Smoke1,
			args: args{key: Smoke1},
			want: toolkit.Want(domain.ValType{"hello": "world", "age": float64(42)}, nil),
		},
		{
			name: Smoke2,
			args: args{key: ""},
			want: toolkit.Want[domain.ValType](nil, domain.ErrEmptyKey),
		},
		{
			name: Smoke3,
			args: args{key: Smoke3},
			want: toolkit.Want[domain.ValType](nil, errDummy),
		},
		{
			name: Smoke4,
			args: args{key: Smoke4},
			want: toolkit.Want[domain.ValType](nil, domain.ErrDataCorrupted),
		},
	}

	useCase := domain.NewGetUseCase(getter{})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			got, err := useCase.Execute(ctx, test.args.key)

			toolkit.Assert(t, toolkit.Got(err, got), test.want)
		})
	}
}

func TestUnitSetUseCase(t *testing.T) {
	t.Parallel()

	type args struct {
		key string
		val domain.ValType
		ttl int
	}

	tests := []struct {
		name string
		args args
		want toolkit.W[any]
	}{
		{
			name: Smoke1,
			args: args{key: Smoke1, val: domain.ValType{"hello": "world", "age": 42}, ttl: 0},
			want: toolkit.Err(nil),
		},
		{
			name: Smoke2,
			args: args{key: Smoke2, val: domain.ValType{"hello": "world", "age": 42}, ttl: 10},
			want: toolkit.Err(nil),
		},
		{
			name: Smoke3,
			args: args{key: Smoke3, val: domain.ValType{"hello": "world", "age": 42}, ttl: -10},
			want: toolkit.Err(nil),
		},
		{
			name: Smoke4,
			args: args{key: "", val: domain.ValType{"hello": "world", "age": 42}, ttl: 0},
			want: toolkit.Err(domain.ErrEmptyKey),
		},
		{name: Smoke5, args: args{key: Smoke5, val: nil, ttl: 0}, want: toolkit.Err(domain.ErrEmptyVal)},
		{
			name: Smoke6,
			args: args{key: Smoke6, val: domain.ValType{"hello": cannotMarshal{}}, ttl: 0},
			want: toolkit.Err(domain.ErrDataCorrupted),
		},
		{
			name: Smoke7,
			args: args{key: Smoke7, val: domain.ValType{"hello": "world", "age": 42}, ttl: 0},
			want: toolkit.Err(errDummy),
		},
	}

	useCase := domain.NewSetUseCase(setter{})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			err := useCase.Execute(ctx, test.args.key, test.args.val, test.args.ttl)

			toolkit.Assert(t, toolkit.Got[any](err), test.want)
		})
	}
}

func TestUnitDelUseCase(t *testing.T) {
	t.Parallel()

	type args struct {
		key string
	}

	tests := []struct {
		name string
		args args
		want toolkit.W[any]
	}{
		{name: Smoke1, args: args{key: Smoke1}, want: toolkit.Err(nil)},
		{name: Smoke2, args: args{key: ""}, want: toolkit.Err(domain.ErrEmptyKey)},
		{name: Smoke3, args: args{key: Smoke3}, want: toolkit.Err(errDummy)},
	}

	useCase := domain.NewDelUseCase(deleter{})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			err := useCase.Execute(ctx, test.args.key)

			toolkit.Assert(t, toolkit.Got[any](err), test.want)
		})
	}
}
