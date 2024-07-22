package blender_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kxnes/go-interviews/apicache/pkg/blender"
	"github.com/kxnes/go-interviews/apicache/test/toolkit"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type (
	json struct {
		Name string `json:"name" validate:"required"`
		ID   int    `json:"id"`
		Path string `json:"-"    param:"path"`
	}
	path struct {
		Name string `param:"name" validate:"required"`
		ID   int    `param:"id"`
		JSON string `json:"json"`
	}
)

var (
	errJSON = errors.New(
		"json error: " +
			"code=400, message=Unmarshal type error: expected=int, got=string, field=id, offset=27, " +
			"internal=json: cannot unmarshal string into Go struct field json.id of type int",
	)
	errPath = errors.New(
		"path error: " +
			`code=400, message=strconv.ParseInt: parsing "six": invalid syntax, ` +
			`internal=strconv.ParseInt: parsing "six": invalid syntax`,
	)
	errValidateJSON = errors.New(
		"validate error: " +
			"Key: 'json.Name' Error:Field validation for 'Name' failed on the 'required' tag",
	)
	errValidatePath = errors.New(
		"validate error: " +
			"Key: 'path.Name' Error:Field validation for 'Name' failed on the 'required' tag",
	)
)

func TestUnitNew(t *testing.T) {
	t.Parallel()

	obj := blender.New[int]()

	assert.IsType(t, new(blender.Blender[int]), obj)
}

func TestUnitBlenderJSON(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string
		body string
		want toolkit.W[*json]
	}

	tests := []testCase{
		{
			name: "success",
			body: `{"name":"Kirill","id":666,"path":"ignore"}`,
			want: toolkit.Want(&json{Name: "Kirill", ID: 666, Path: ""}, nil),
		},
		{
			name: "json error",
			body: `{"name":"Kirill","id":"666"}`,
			want: toolkit.Want[*json](nil, errJSON),
		},
		{
			name: "validate error",
			body: `{"id":666}`,
			want: toolkit.Want[*json](nil, errValidateJSON),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			obj := blender.New[json]()
			req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(test.body))
			rec := httptest.NewRecorder()

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			got, err := obj.JSON(echo.New().NewContext(req, rec))

			toolkit.Assert(t, toolkit.Got(err, got), test.want)
		})
	}
}

func TestUnitBlenderPath(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string
		path []string
		want toolkit.W[*path]
	}

	tests := []testCase{
		{
			name: "success",
			path: []string{"Kirill", "666"},
			want: toolkit.Want(&path{Name: "Kirill", ID: 666, JSON: ""}, nil),
		},
		{
			name: "path error",
			path: []string{"Kirill", "six"},
			want: toolkit.Want[*path](nil, errPath),
		},
		{
			name: "validate error",
			path: []string{"", "666"},
			want: toolkit.Want[*path](nil, errValidatePath),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			obj := blender.New[path]()
			req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(`{"json"":"ignore"}`))
			rec := httptest.NewRecorder()
			etx := echo.New().NewContext(req, rec)

			etx.SetParamNames("name", "id")
			etx.SetParamValues(test.path...)

			got, err := obj.Path(etx)

			toolkit.Assert(t, toolkit.Got(err, got), test.want)
		})
	}
}
