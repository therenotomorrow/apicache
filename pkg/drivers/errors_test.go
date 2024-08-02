package drivers_test

import (
	"testing"

	"github.com/kxnes/go-interviews/apicache/pkg/drivers"
	"github.com/kxnes/go-interviews/apicache/test/toolkit"
)

func TestUnitErrNotExist(t *testing.T) {
	t.Parallel()

	toolkit.Assert(t, toolkit.Got(nil, drivers.ErrNotExist.Error()), toolkit.Want("entity not exist", nil))
}
