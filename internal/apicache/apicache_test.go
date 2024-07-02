package apicache

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/kxnes/go-interviews/apicache/internal/fs"
	"github.com/kxnes/go-interviews/apicache/test"
)

func TestServer(t *testing.T) {
	dummyReq := func(addr string, r chan int) {
		resp, err := http.Get(addr)

		if err != nil {
			r <- 0
			return
		}

		r <- resp.StatusCode
	}

	dm := &test.DriverMock{
		Storage:      &sync.Map{},
		IsConcurrent: true,
	}

	srv := NewServer(
		&Dependencies{Driver: fs.New(dm, &fs.Options{
			MaxConn: 0,
			Timeout: 1,
		})},
		&Options{Addr: "127.0.0.1:8080"},
	)

	go srv.Listen()

	// Wait for server start
	time.Sleep(2 * time.Second)

	responses := make(chan int, 2)

	go dummyReq("http://"+srv.opts.Addr+"/"+test.KeyError, responses)
	go dummyReq("http://"+srv.opts.Addr+"/"+test.KeyError, responses)

	resp1 := <-responses
	resp2 := <-responses

	if !(resp1 == 408 && resp2 == 500 || resp2 == 408 && resp1 != 500) {
		t.Errorf("responses not match: %d vs. %d and %d vs. %d", resp1, 408, resp2, 500)
	}

	err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)

	if err != nil {
		t.Errorf("syscall.Kill() unexpected error = %v", err)
	}

	if _, ok := <-srv.done; ok {
		t.Errorf("server not closed")
	}

	if dm.Closed != 1 {
		t.Errorf("dependent driver not closed")
	}
}

func TestServerErrors(t *testing.T) {
	type Case struct {
		name string
		deps *fs.Options
		opts *Options
		err  string
	}

	cases := []Case{
		{
			name: "negative maxConn",
			deps: &fs.Options{
				MaxConn: -maxConn,
				Timeout: timeout,
			},
			opts: &Options{Addr: "127.0.0.1:8080"},
			err:  "non-positive MaxConn",
		},
		{
			name: "negative timeout",
			deps: &fs.Options{
				MaxConn: 0,
				Timeout: -timeout,
			},
			opts: &Options{Addr: "127.0.0.1:8080"},
			err:  "non-positive Timeout",
		},
		{
			name: "wrong address",
			deps: &fs.Options{
				MaxConn: maxConn,
				Timeout: timeout,
			},
			opts: &Options{Addr: "127.0.0.666:8080"},
			err:  "server listen err = listen tcp: lookup 127.0.0.666: no such host",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			go func(c Case) {
				defer func() {
					err := recover()
					if err.(string) != c.err {
						t.Errorf("panic got = %v, want = %v", err, c.err)
					}
				}()

				NewServer(&Dependencies{Driver: fs.New(&test.DriverMock{
					Storage:      &sync.Map{},
					IsConcurrent: true,
				}, c.deps)}, c.opts).Listen()
			}(c)

			// Wait for server start
			time.Sleep(2 * time.Second)
		})
	}
}

const (
	keyExist = "exist"
	valExist = "exist"
	ttlExist = 10
	maxConn  = 10
	timeout  = 10
)

var d = fs.New(
	&test.DriverMock{
		Storage: &sync.Map{},
	},
	&fs.Options{
		MaxConn: maxConn,
		Timeout: timeout,
	},
)

func TestStorageHandlerGet(t *testing.T) {
	_ = d.Set(keyExist, valExist, ttlExist)
	ts := httptest.NewServer(&StorageHandler{driver: d})
	defer ts.Close()

	cases := []struct {
		name string
		key  string
		code int
		body string
	}{
		{
			name: "500 (internal driver error)",
			key:  "/" + test.KeyError,
			code: http.StatusInternalServerError,
			body: `{"error":"internal error"}`,
		},
		{
			name: "404",
			key:  "/" + test.KeyNotExist,
			code: http.StatusNotFound,
			body: `{"error":"key (` + test.KeyNotExist + `) not exist"}`,
		},
		{
			name: "400",
			key:  "",
			code: http.StatusBadRequest,
			body: `{"error":"empty key"}`,
		},
		{
			name: "200",
			key:  "/" + keyExist,
			code: http.StatusOK,
			body: `{"value":"` + keyExist + `"}`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Get(ts.URL + c.key)
			if err != nil {
				t.Errorf("GET unexpected error = %v", err)
				return
			}

			if resp.StatusCode != c.code {
				t.Errorf("GET code = %v, want = %v", resp.StatusCode, c.code)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("GET unexpected body read = %v", err)
				return
			}
			defer func() { _ = resp.Body.Close() }()

			got := strings.TrimSpace(string(body))
			if got != c.body {
				t.Errorf("GET body = %v, want = %v", got, c.body)
			}
		})
	}
}

func TestStorageHandlerPost(t *testing.T) {
	ts := httptest.NewServer(&StorageHandler{driver: d})
	defer ts.Close()

	cases := []struct {
		name string
		code int
		body string
		form string
	}{
		{
			name: "400 (invalid type)",
			code: http.StatusBadRequest,
			body: `{"error":"invalid type (int) for field (ttl)"}`,
			form: `{"key":"1","val":"1","ttl":"1"}`,
		},
		{
			name: "400 (invalid JSON)",
			code: http.StatusBadRequest,
			body: `{"error":"invalid JSON"}`,
			form: `{`,
		},
		{
			name: "400 (invalid field)",
			code: http.StatusBadRequest,
			body: `{"error":"empty value for key (error)"}`,
			form: fmt.Sprintf(`{"key":"%s","val":"","ttl":1}`, test.KeyError),
		},
		{
			name: "500 (internal server error)",
			code: http.StatusInternalServerError,
			body: `{"error":"internal error"}`,
			form: fmt.Sprintf(`{"key":"%s","val":"1","ttl":1}`, test.KeyError),
		},
		{
			name: "201",
			code: http.StatusCreated,
			body: "",
			form: fmt.Sprintf(`{"key":"%s","val":"1","ttl":1}`, keyExist),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp, err := http.Post(ts.URL, "application/json", strings.NewReader(c.form))
			if err != nil {
				t.Errorf("POST unexpected error = %v", err)
				return
			}

			if resp.StatusCode != c.code {
				t.Errorf("POST code = %v, want = %v", resp.StatusCode, c.code)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("POST unexpected body read = %v", err)
				return
			}
			defer func() { _ = resp.Body.Close() }()

			got := strings.TrimSpace(string(body))
			if got != c.body {
				t.Errorf("POST body = %v, want = %v", got, c.body)
			}
		})
	}
}

func TestStorageHandlerDelete(t *testing.T) {
	ts := httptest.NewServer(&StorageHandler{driver: d})
	defer ts.Close()

	cases := []struct {
		name string
		key  string
		code int
		body string
	}{
		{
			name: "500 (internal driver error)",
			key:  "/" + test.KeyError,
			code: http.StatusInternalServerError,
			body: `{"error":"internal error"}`,
		},
		{
			name: "404",
			key:  "/" + test.KeyNotExist,
			code: http.StatusNotFound,
			body: `{"error":"key (` + test.KeyNotExist + `) not exist"}`,
		},
		{
			name: "400",
			key:  "",
			code: http.StatusBadRequest,
			body: `{"error":"empty key"}`,
		},
		{
			name: "204",
			key:  "/" + keyExist,
			code: http.StatusNoContent,
			body: ``,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, ts.URL+c.key, nil)
			if err != nil {
				t.Errorf("DELETE new request unexpected error = %v", err)
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("DELETE unexpected error = %v", err)
				return
			}

			if resp.StatusCode != c.code {
				t.Errorf("DELETE code = %v, want = %v", resp.StatusCode, c.code)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("DELETE unexpected body read = %v", err)
				return
			}
			defer func() { _ = resp.Body.Close() }()

			got := strings.TrimSpace(string(body))
			if got != c.body {
				t.Errorf("DELETE body = %v, want = %v", got, c.body)
			}
		})
	}
}

func TestStorageHandlerMethodNotAllowed(t *testing.T) {
	ts := httptest.NewServer(&StorageHandler{driver: d})
	defer ts.Close()

	cases := []struct {
		name string
		key  string
		code int
		body string
	}{
		{
			name: "405",
			key:  "/" + test.KeyError,
			code: http.StatusMethodNotAllowed,
			body: `{}`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPut, ts.URL+c.key, nil)
			if err != nil {
				t.Errorf("PUT new request unexpected error = %v", err)
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("PUT unexpected error = %v", err)
				return
			}

			if resp.StatusCode != c.code {
				t.Errorf("PUT code = %v, want = %v", resp.StatusCode, c.code)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("PUT unexpected body read = %v", err)
				return
			}
			defer func() { _ = resp.Body.Close() }()

			got := strings.TrimSpace(string(body))
			if got != c.body {
				t.Errorf("PUT body = %v, want = %v", got, c.body)
			}
		})
	}
}

func TestStorageHandlersRequestTimeout(t *testing.T) {
	d := fs.New(
		&test.DriverMock{
			Storage:      &sync.Map{},
			IsConcurrent: true,
		},
		&fs.Options{
			MaxConn: 0,
			Timeout: 1,
		},
	)
	ts := httptest.NewServer(&StorageHandler{driver: d})
	defer ts.Close()

	cases := []struct {
		name   string
		key    string
		method string
		code   int
		body   string
		form   string
	}{
		{
			name:   "408 (GET)",
			key:    "/" + test.KeyError,
			method: http.MethodGet,
			code:   http.StatusRequestTimeout,
			body:   `{"error":"timeout for operation (get)"}`,
		},
		{
			name:   "408 (POST)",
			key:    "/",
			method: http.MethodPost,
			code:   http.StatusRequestTimeout,
			body:   `{"error":"timeout for operation (set)"}`,
			form:   `{"key":"1","val":"1","ttl":1}`,
		},
		{
			name:   "408 (DELETE)",
			method: http.MethodDelete,
			key:    "/" + test.KeyError,
			code:   http.StatusRequestTimeout,
			body:   `{"error":"timeout for operation (del)"}`,
		},
		{
			name:   "400 (non-blocking parse JSON 1)",
			method: http.MethodPost,
			code:   http.StatusBadRequest,
			body:   `{"error":"invalid type (int) for field (ttl)"}`,
			form:   `{"key":"1","val":"1","ttl":"1"}`,
		},
		{
			name:   "400 (non-blocking parse JSON 2)",
			method: http.MethodPost,
			code:   http.StatusBadRequest,
			body:   `{"error":"invalid JSON"}`,
			form:   `{`,
		},
		{
			name:   "405 (non-blocking HTTP request)",
			method: http.MethodPatch,
			key:    "/" + test.KeyError,
			code:   http.StatusMethodNotAllowed,
			body:   `{}`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			go func() { _, _ = http.Get(ts.URL + "/" + test.KeyError) }()

			req, err := http.NewRequest(c.method, ts.URL+c.key, strings.NewReader(c.form))
			if err != nil {
				t.Errorf("%s new request unexpected error = %v", strings.ToUpper(c.method), err)
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("%s unexpected error = %v", strings.ToUpper(c.method), err)
				return
			}

			if resp.StatusCode != c.code {
				t.Errorf("%s code = %v, want = %v", strings.ToUpper(c.method), resp.StatusCode, c.code)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("%s unexpected body read = %v", strings.ToUpper(c.method), err)
				return
			}
			defer func() { _ = resp.Body.Close() }()

			got := strings.TrimSpace(string(body))
			if got != c.body {
				t.Errorf("%s body = %v, want = %v", strings.ToUpper(c.method), got, c.body)
			}
		})
	}
}
