package apiv1get_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	apiv1get "github.com/therenotomorrow/apicache/internal/api/v1/get"
	"github.com/therenotomorrow/apicache/internal/domain"
	"github.com/therenotomorrow/apicache/test/toolkit"
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
	cacheGetter struct{}
	params      struct {
		names  []string
		values []string
	}
	args struct {
		params *params
	}
	want struct {
		code int
		body string
	}
	testCase struct {
		name string
		args args
		want want
	}
)

func (c cacheGetter) Get(_ context.Context, key string) ([]byte, error) {
	switch key {
	case Smoke2:
		return nil, domain.ErrKeyExpired
	case Smoke3:
		return nil, domain.ErrKeyNotExist
	case Smoke4:
		return nil, domain.ErrConnTimeout
	case Smoke5:
		return nil, domain.ErrContextTimeout
	case Smoke6:
		return nil, errDummy
	}

	return []byte(`{"hello":"world","age":42}`), nil
}

func successTC() testCase {
	return testCase{
		name: Smoke1,
		args: args{params: &params{names: []string{"key"}, values: []string{Smoke1}}},
		want: want{code: http.StatusOK, body: `{"key":"smoke1","val":{"age":42,"hello":"world"}}`},
	}
}

func expiredKeyTC() testCase {
	return testCase{
		name: Smoke2,
		args: args{params: &params{names: []string{"key"}, values: []string{Smoke2}}},
		want: want{code: http.StatusBadRequest, body: `{"message":"key is expired"}`},
	}
}

func keyNotExistTC() testCase {
	return testCase{
		name: Smoke3,
		args: args{params: &params{names: []string{"key"}, values: []string{Smoke3}}},
		want: want{code: http.StatusNotFound, body: `{"message":"key not exist"}`},
	}
}

func connectionTimeoutTC() testCase {
	return testCase{
		name: Smoke4,
		args: args{params: &params{names: []string{"key"}, values: []string{Smoke4}}},
		want: want{code: http.StatusTooManyRequests, body: `{"message":"connection timeout"}`},
	}
}

func contextTimeoutTC() testCase {
	return testCase{
		name: Smoke5,
		args: args{params: &params{names: []string{"key"}, values: []string{Smoke5}}},
		want: want{code: http.StatusTooManyRequests, body: `{"message":"context timeout"}`},
	}
}

func failureTC() testCase {
	return testCase{
		name: Smoke6,
		args: args{params: &params{names: []string{"key"}, values: []string{Smoke6}}},
		want: want{code: http.StatusInternalServerError, body: `{"message":"InternalServerError"}`},
	}
}

func invalidParamsTC() testCase {
	return testCase{
		name: Smoke7,
		args: args{params: &params{names: []string{"key"}, values: nil}},
		want: want{
			code: http.StatusUnprocessableEntity,
			body: "{\"message\":\"validate error: Key: 'Params.Key' Error:" +
				"Field validation for 'Key' failed on the 'required' tag\"}",
		},
	}
}

func TestUnitGet(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		successTC(),
		expiredKeyTC(),
		keyNotExistTC(),
		connectionTimeoutTC(),
		contextTimeoutTC(),
		failureTC(),
		invalidParamsTC(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			mux := echo.New()

			etx := mux.NewContext(req, rec)
			etx.SetParamNames(test.args.params.names...)
			etx.SetParamValues(test.args.params.values...)

			mux.HTTPErrorHandler(apiv1get.Get(cacheGetter{})(etx), etx)

			toolkit.Assert(t, toolkit.Got(nil, rec.Code), toolkit.Want(test.want.code, nil))
			toolkit.Assert(t, toolkit.Got(nil, strings.TrimSpace(rec.Body.String())), toolkit.Want(test.want.body, nil))
		})
	}
}
