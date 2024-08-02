package api_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kxnes/go-interviews/apicache/internal/api"
	"github.com/kxnes/go-interviews/apicache/test/toolkit"
	"github.com/labstack/echo/v4"
)

const (
	regularMsg = "{\"message\":\"dummy error\"}"
	debugMsg   = "{\n  \"message\": \"dummy error\"\n}"
	emptyMsg   = "null"
)

var errDummy = errors.New("dummy error")

type (
	args struct {
		err   error
		debug bool
	}
	testCase struct {
		name string
		args args
		want toolkit.W[string]
	}
)

func handleError(herr *echo.HTTPError, debug bool) (int, string) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mux := echo.New()
	mux.Debug = debug
	etx := mux.NewContext(req, rec)

	mux.HTTPErrorHandler(herr, etx)

	return rec.Code, strings.TrimSpace(rec.Body.String())
}

func testCases() []testCase {
	return []testCase{
		{name: "with error", args: args{err: errDummy, debug: false}, want: toolkit.Want(regularMsg, nil)},
		{name: "with debug", args: args{err: errDummy, debug: true}, want: toolkit.Want(debugMsg, nil)},
		{name: "without error", args: args{err: nil, debug: false}, want: toolkit.Want(emptyMsg, nil)},
	}
}

func TestUnitBadRequestError(t *testing.T) {
	t.Parallel()

	for _, test := range testCases() {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			code, body := handleError(api.BadRequestError(test.args.err), test.args.debug)

			toolkit.Assert(t, toolkit.Got(nil, code), toolkit.Want(http.StatusBadRequest, nil))
			toolkit.Assert(t, toolkit.Got(nil, body), test.want)
		})
	}
}

func TestUnitNotFoundError(t *testing.T) {
	t.Parallel()

	for _, test := range testCases() {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			code, body := handleError(api.NotFoundError(test.args.err), test.args.debug)

			toolkit.Assert(t, toolkit.Got(nil, code), toolkit.Want(http.StatusNotFound, nil))
			toolkit.Assert(t, toolkit.Got(nil, body), test.want)
		})
	}
}

func TestUnitUnprocessableEntityError(t *testing.T) {
	t.Parallel()

	for _, test := range testCases() {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			code, body := handleError(api.UnprocessableEntityError(test.args.err), test.args.debug)

			toolkit.Assert(t, toolkit.Got(nil, code), toolkit.Want(http.StatusUnprocessableEntity, nil))
			toolkit.Assert(t, toolkit.Got(nil, body), test.want)
		})
	}
}

func TestUnitTooManyRequestsError(t *testing.T) {
	t.Parallel()

	for _, test := range testCases() {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			code, body := handleError(api.TooManyRequestsError(test.args.err), test.args.debug)

			toolkit.Assert(t, toolkit.Got(nil, code), toolkit.Want(http.StatusTooManyRequests, nil))
			toolkit.Assert(t, toolkit.Got(nil, body), test.want)
		})
	}
}

func TestUnitInternalServerError(t *testing.T) {
	t.Parallel()

	type args struct {
		err     error
		debug   bool
		message []string
	}

	tests := []struct {
		name string
		args args
		want toolkit.W[string]
	}{
		{
			name: "with error",
			args: args{err: errDummy, message: []string{}, debug: false},
			want: toolkit.Want(`{"message":"InternalServerError"}`, nil),
		},
		{
			name: "without error",
			args: args{err: nil, message: []string{}, debug: false},
			want: toolkit.Want(`{"message":"InternalServerError"}`, nil),
		},
		{
			name: "custom message",
			args: args{err: nil, message: []string{"dummy", "error"}, debug: false},
			want: toolkit.Want(`{"message":"dummy"}`, nil),
		},
		{
			name: "with debug",
			args: args{err: errDummy, message: []string{"dummy", "error"}, debug: true},
			want: toolkit.Want(
				"{\n  \"error\": \"code=500, message=dummy, internal=dummy error\",\n  \"message\": \"dummy\"\n}",
				nil,
			),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			code, body := handleError(api.InternalServerError(test.args.err, test.args.message...), test.args.debug)

			toolkit.Assert(t, toolkit.Got(nil, code), toolkit.Want(http.StatusInternalServerError, nil))
			toolkit.Assert(t, toolkit.Got(nil, body), test.want)
		})
	}
}
