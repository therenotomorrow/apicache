package toolkit

import (
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type G[T any] struct {
	got T
	err error
}

type W[T any] struct {
	want T
	err  error
}

func Got[T any](err error, got ...T) G[T] {
	if len(got) > 0 {
		return G[T]{got: got[0], err: err}
	}

	var def T

	return G[T]{got: def, err: err}
}

func Want[T any](want T, err error) W[T] {
	return W[T]{want: want, err: err}
}

func Err(err error) W[any] {
	return W[any]{want: nil, err: err}
}

func failure[T any](t *testing.T, got T, err error, want error) {
	t.Helper()

	assert.Empty(t, got)

	require.EqualError(t, err, want.Error())
}

func success[T any](t *testing.T, got T, err error, want T) {
	t.Helper()

	require.NoError(t, err)

	assert.Equal(t, want, got)
}

func Assert[T any](t *testing.T, got G[T], want W[T]) {
	t.Helper()

	if want.err != nil {
		failure(t, got.got, got.err, want.err)
	} else {
		success(t, got.got, got.err, want.want)
	}
}

func RootDir() string {
	_, filename, _, ok := runtime.Caller(0)

	if !ok {
		return ""
	}

	return path.Join(path.Dir(filename), "..", "..")
}

func EnvFile() string {
	return path.Join(RootDir(), "configs", ".env.test")
}
