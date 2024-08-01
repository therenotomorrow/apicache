package apiv1delete_test

import (
	"context"
	"errors"
	apiv1delete "github.com/kxnes/go-interviews/apicache/internal/api/v1/delete"
	"github.com/kxnes/go-interviews/apicache/pkg/cache"
	"github.com/kxnes/go-interviews/apicache/test/toolkit"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	Smoke1 = "smoke1"
	Smoke2 = "smoke2"
	Smoke3 = "smoke3"
	Smoke4 = "smoke4"
	Smoke5 = "smoke5"
)

var errDummy = errors.New("dummy error")

type cacheDeleter struct{}

func (c cacheDeleter) Del(_ context.Context, key string) error {
	switch key {
	case Smoke2:
		return cache.ErrConnTimeout
	case Smoke3:
		return cache.ErrContextTimeout
	case Smoke4:
		return errDummy
	default:
		return nil
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	type params struct {
		names  []string
		values []string
	}

	type args struct {
		params *params
	}

	type want struct {
		code int
		body string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: Smoke1,
			args: args{params: &params{names: []string{"key"}, values: []string{Smoke1}}},
			want: want{code: http.StatusNoContent, body: ""},
		},
		{
			name: Smoke2,
			args: args{params: &params{names: []string{"key"}, values: []string{Smoke2}}},
			want: want{code: http.StatusTooManyRequests, body: `{"message":"connection timeout"}`},
		},
		{
			name: Smoke3,
			args: args{params: &params{names: []string{"key"}, values: []string{Smoke3}}},
			want: want{code: http.StatusTooManyRequests, body: `{"message":"context timeout"}`},
		},
		{
			name: Smoke4,
			args: args{params: &params{names: []string{"key"}, values: []string{Smoke4}}},
			want: want{code: http.StatusInternalServerError, body: `{"message":"InternalServerError"}`},
		},
		{
			name: Smoke5,
			args: args{params: &params{names: []string{"key"}, values: nil}},
			want: want{code: http.StatusUnprocessableEntity, body: "{\"message\":\"validate error: Key: 'Params.Key' Error:Field validation for 'Key' failed on the 'required' tag\"}"},
		},
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
