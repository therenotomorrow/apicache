package drivers_test

import (
	"testing"

	"github.com/therenotomorrow/apicache/pkg/drivers"
	"github.com/therenotomorrow/apicache/test/toolkit"
)

func TestUnitErrNotExist(t *testing.T) {
	t.Parallel()

	toolkit.Assert(t, toolkit.Got(nil, drivers.ErrNotExist.Error()), toolkit.Want("entity not exist", nil))
}
