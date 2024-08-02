package domain_test

import (
	"testing"

	"github.com/kxnes/go-interviews/apicache/internal/domain"
	"github.com/kxnes/go-interviews/apicache/test/toolkit"
)

func TestUnitErrKeyNotExist(t *testing.T) {
	t.Parallel()

	toolkit.Assert(t, toolkit.Got(nil, domain.ErrKeyNotExist.Error()), toolkit.Want("key not exist", nil))
}

func TestUnitErrKeyExpired(t *testing.T) {
	t.Parallel()

	toolkit.Assert(t, toolkit.Got(nil, domain.ErrKeyExpired.Error()), toolkit.Want("key is expired", nil))
}

func TestUnitErrConnTimeout(t *testing.T) {
	t.Parallel()

	toolkit.Assert(t, toolkit.Got(nil, domain.ErrConnTimeout.Error()), toolkit.Want("connection timeout", nil))
}

func TestUnitErrContextTimeout(t *testing.T) {
	t.Parallel()

	toolkit.Assert(t, toolkit.Got(nil, domain.ErrContextTimeout.Error()), toolkit.Want("context timeout", nil))
}

func TestUnitErrClosed(t *testing.T) {
	t.Parallel()

	toolkit.Assert(t, toolkit.Got(nil, domain.ErrClosed.Error()), toolkit.Want("closed instance", nil))
}

func TestUnitErrEmptyKey(t *testing.T) {
	t.Parallel()

	toolkit.Assert(t, toolkit.Got(nil, domain.ErrEmptyKey.Error()), toolkit.Want("empty key", nil))
}

func TestUnitErrEmptyVal(t *testing.T) {
	t.Parallel()

	toolkit.Assert(t, toolkit.Got(nil, domain.ErrEmptyVal.Error()), toolkit.Want("empty value", nil))
}
