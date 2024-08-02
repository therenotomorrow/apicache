package apiv1delete_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apiv1delete "github.com/kxnes/go-interviews/apicache/internal/api/v1/delete"
	"github.com/kxnes/go-interviews/apicache/internal/domain"
	"github.com/kxnes/go-interviews/apicache/test/toolkit"
	"github.com/labstack/echo/v4"
)

const (
	Smoke1 = "smoke1"
	Smoke2 = "smoke2"
	Smoke3 = "smoke3"
	Smoke4 = "smoke4"
	Smoke5 = "smoke5"
)

var errDummy = errors.New("dummy error")

type (
	cacheDeleter struct{}
	params       struct {
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

func (c cacheDeleter) Del(_ context.Context, key string) error {
	switch key {
	case Smoke2:
		return domain.ErrConnTimeout
	case Smoke3:
		return domain.ErrContextTimeout
	case Smoke4:
		return errDummy
	}

	return nil
}

func successTC() testCase {
	return testCase{
		name: Smoke1,
		args: args{params: &params{names: []string{"key"}, values: []string{Smoke1}}},
		want: want{code: http.StatusNoContent, body: ""},
	}
}

func connectionTimeoutTC() testCase {
	return testCase{
		name: Smoke2,
		args: args{params: &params{names: []string{"key"}, values: []string{Smoke2}}},
		want: want{code: http.StatusTooManyRequests, body: `{"message":"connection timeout"}`},
	}
}

func contextTimeoutTC() testCase {
	return testCase{
		name: Smoke3,
		args: args{params: &params{names: []string{"key"}, values: []string{Smoke3}}},
		want: want{code: http.StatusTooManyRequests, body: `{"message":"context timeout"}`},
	}
}

func failureTC() testCase {
	return testCase{
		name: Smoke4,
		args: args{params: &params{names: []string{"key"}, values: []string{Smoke4}}},
		want: want{code: http.StatusInternalServerError, body: `{"message":"InternalServerError"}`},
	}
}

func invalidParamsTC() testCase {
	return testCase{
		name: Smoke5,
		args: args{params: &params{names: []string{"key"}, values: nil}},
		want: want{
			code: http.StatusUnprocessableEntity,
			body: "{\"message\":\"validate error: Key: 'Params.Key' Error:" +
				"Field validation for 'Key' failed on the 'required' tag\"}",
		},
	}
}

func TestUnitDelete(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		successTC(),
		connectionTimeoutTC(),
		contextTimeoutTC(),
		failureTC(),
		invalidParamsTC(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodDelete, "/", nil)
			rec := httptest.NewRecorder()
			mux := echo.New()

			etx := mux.NewContext(req, rec)
			etx.SetParamNames(test.args.params.names...)
			etx.SetParamValues(test.args.params.values...)

			mux.HTTPErrorHandler(apiv1delete.Delete(cacheDeleter{})(etx), etx)

			toolkit.Assert(t, toolkit.Got(nil, rec.Code), toolkit.Want(test.want.code, nil))
			toolkit.Assert(t, toolkit.Got(nil, strings.TrimSpace(rec.Body.String())), toolkit.Want(test.want.body, nil))
		})
	}
}
