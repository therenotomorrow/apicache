package apiv1post_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	apiv1post "github.com/therenotomorrow/apicache/internal/api/v1/post"
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
	Smoke8 = "smoke8"
)

var errDummy = errors.New("dummy error")

type (
	cacheSetter struct{}
	params      struct {
		names  []string
		values []string
	}
	args struct {
		params  *params
		payload string
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

func (c cacheSetter) Set(_ context.Context, key string, _ []byte, _ time.Time) error {
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
		args: args{
			params:  &params{names: []string{"key"}, values: []string{Smoke1}},
			payload: `{"val":{"age":42,"hello":"world"},"ttl":10}`,
		},
		want: want{code: http.StatusCreated, body: `{"key":"smoke1","val":{"age":42,"hello":"world"}}`},
	}
}

func connectionTimeoutTC() testCase {
	return testCase{
		name: Smoke2,
		args: args{
			params:  &params{names: []string{"key"}, values: []string{Smoke2}},
			payload: `{"val":{"age":42,"hello":"world"},"ttl":10}`,
		},
		want: want{code: http.StatusTooManyRequests, body: `{"message":"connection timeout"}`},
	}
}

func contextTimeoutTC() testCase {
	return testCase{
		name: Smoke3,
		args: args{
			params:  &params{names: []string{"key"}, values: []string{Smoke3}},
			payload: `{"val":{"age":42,"hello":"world"},"ttl":10}`,
		},
		want: want{code: http.StatusTooManyRequests, body: `{"message":"context timeout"}`},
	}
}

func failureTC() testCase {
	return testCase{
		name: Smoke4,
		args: args{
			params:  &params{names: []string{"key"}, values: []string{Smoke4}},
			payload: `{"val":{"age":42,"hello":"world"},"ttl":10}`,
		},
		want: want{code: http.StatusInternalServerError, body: `{"message":"InternalServerError"}`},
	}
}

func invalidParamsTC() testCase {
	return testCase{
		name: Smoke5,
		args: args{
			params:  &params{names: []string{"key"}, values: nil},
			payload: "",
		},
		want: want{
			code: http.StatusUnprocessableEntity,
			body: "{\"message\":\"validate error: Key: 'Params.Key' Error:" +
				"Field validation for 'Key' failed on the 'required' tag\"}",
		},
	}
}

func nonRequiredPayloadTC() testCase {
	return testCase{
		name: Smoke6,
		args: args{
			params:  &params{names: []string{"key"}, values: []string{Smoke6}},
			payload: `{"val":{"age":42,"hello":"world"}}`,
		},
		want: want{code: http.StatusCreated, body: `{"key":"smoke6","val":{"age":42,"hello":"world"}}`},
	}
}

func ttlMinValidationTC() testCase {
	return testCase{
		name: Smoke7,
		args: args{
			params:  &params{names: []string{"key"}, values: []string{Smoke7}},
			payload: `{"val":{"age":42,"hello":"world"},"ttl":-10}`,
		},
		want: want{
			code: http.StatusUnprocessableEntity,
			body: "{\"message\":\"validate error: Key: 'Payload.TTL' Error:" +
				"Field validation for 'TTL' failed on the 'min' tag\"}",
		},
	}
}

func requiredPayloadTC() testCase {
	return testCase{
		name: Smoke8,
		args: args{
			params:  &params{names: []string{"key"}, values: []string{Smoke8}},
			payload: `{"ttl":10}`,
		},
		want: want{
			code: http.StatusUnprocessableEntity,
			body: "{\"message\":\"validate error: Key: 'Payload.Val' Error:" +
				"Field validation for 'Val' failed on the 'required' tag\"}",
		},
	}
}

func TestUnitPost(t *testing.T) {
	t.Parallel()

	tests := []testCase{
		successTC(),
		connectionTimeoutTC(),
		contextTimeoutTC(),
		failureTC(),
		invalidParamsTC(),
		nonRequiredPayloadTC(),
		requiredPayloadTC(),
		ttlMinValidationTC(),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.args.payload))
			rec := httptest.NewRecorder()
			mux := echo.New()

			req.Header.Set("Content-Type", "application/json")

			etx := mux.NewContext(req, rec)
			etx.SetParamNames(test.args.params.names...)
			etx.SetParamValues(test.args.params.values...)

			mux.HTTPErrorHandler(apiv1post.Post(cacheSetter{})(etx), etx)

			toolkit.Assert(t, toolkit.Got(nil, rec.Code), toolkit.Want(test.want.code, nil))
			toolkit.Assert(t, toolkit.Got(nil, strings.TrimSpace(rec.Body.String())), toolkit.Want(test.want.body, nil))
		})
	}
}
