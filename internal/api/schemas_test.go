package api_test

import (
	"reflect"
	"testing"

	"github.com/therenotomorrow/apicache/internal/api"
	"github.com/therenotomorrow/apicache/test/toolkit"
)

func TestUnitBadRequest(t *testing.T) {
	t.Parallel()

	tags := reflect.TypeOf(new(api.BadRequest)).Elem().Field(0).Tag

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("enums")),
		toolkit.Want("key is expired", nil),
	)

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("json")),
		toolkit.Want("message", nil),
	)
}

func TestUnitNotFound(t *testing.T) {
	t.Parallel()

	tags := reflect.TypeOf(new(api.NotFound)).Elem().Field(0).Tag

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("enums")),
		toolkit.Want("key not exist", nil),
	)

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("json")),
		toolkit.Want("message", nil),
	)
}

func TestUnitUnprocessableEntity(t *testing.T) {
	t.Parallel()

	tags := reflect.TypeOf(new(api.UnprocessableEntity)).Elem().Field(0).Tag

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("json")),
		toolkit.Want("message", nil),
	)
}

func TestUnitTooManyRequests(t *testing.T) {
	t.Parallel()

	tags := reflect.TypeOf(new(api.TooManyRequests)).Elem().Field(0).Tag

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("enums")),
		toolkit.Want("connection timeout,context timeout", nil),
	)

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("json")),
		toolkit.Want("message", nil),
	)
}

func TestUnitInternalServer(t *testing.T) {
	t.Parallel()

	tags := reflect.TypeOf(new(api.InternalServer)).Elem().Field(0).Tag

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("json")),
		toolkit.Want("message", nil),
	)
}
