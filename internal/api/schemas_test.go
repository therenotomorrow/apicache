package api

import (
	"github.com/kxnes/go-interviews/apicache/test/toolkit"
	"reflect"
	"testing"
)

func TestBadRequest(t *testing.T) {
	tags := reflect.TypeOf(new(BadRequest)).Elem().Field(0).Tag

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("enums")),
		toolkit.Want("key is expired", nil),
	)

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("json")),
		toolkit.Want("message", nil),
	)
}

func TestNotFound(t *testing.T) {
	tags := reflect.TypeOf(new(NotFound)).Elem().Field(0).Tag

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("enums")),
		toolkit.Want("key not exist", nil),
	)

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("json")),
		toolkit.Want("message", nil),
	)
}

func TestUnprocessableEntity(t *testing.T) {
	tags := reflect.TypeOf(new(UnprocessableEntity)).Elem().Field(0).Tag

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("json")),
		toolkit.Want("message", nil),
	)
}

func TestTooManyRequests(t *testing.T) {
	tags := reflect.TypeOf(new(TooManyRequests)).Elem().Field(0).Tag

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("enums")),
		toolkit.Want("connection timeout,context timeout", nil),
	)

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("json")),
		toolkit.Want("message", nil),
	)
}

func TestInternalServer(t *testing.T) {
	tags := reflect.TypeOf(new(InternalServer)).Elem().Field(0).Tag

	toolkit.Assert(t,
		toolkit.Got(nil, tags.Get("json")),
		toolkit.Want("message", nil),
	)
}
