package server_test

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/kxnes/go-interviews/apicache/internal/config"
	"github.com/kxnes/go-interviews/apicache/internal/server"
	"github.com/kxnes/go-interviews/apicache/test/toolkit"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func echoHandler(etx echo.Context) error {
	time.Sleep(time.Minute)

	return etx.String(http.StatusOK, "ok")
}

func TestUnitNew(t *testing.T) {
	t.Parallel()

	type args struct {
		debug bool
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "with debug",
			args: args{debug: true},
		},
		{
			name: "without debug",
			args: args{debug: false},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			settings := config.MustNew(toolkit.EnvFile())
			settings.Debug = test.args.debug

			srv := server.New(settings, nil)

			toolkit.Assert(t, toolkit.Got(nil, *settings), toolkit.Want(srv.Settings(), nil))

			router := srv.UnsafeRouter()

			toolkit.Assert(t, toolkit.Got(nil, settings.Debug), toolkit.Want(router.Debug, nil))
			toolkit.Assert(t, toolkit.Got(nil, log.Lvl(2)), toolkit.Want(router.Logger.Level(), nil))

			existed := make([]string, 0)
			for _, route := range router.Routes() {
				existed = append(existed, route.Method+": "+route.Path)
			}

			expected := []string{
				// ---- cache
				"GET: /api/v1/:key/",
				"POST: /api/v1/:key/",
				"DELETE: /api/v1/:key/",
				// ---- docs
				"GET: /api/docs/*",
			}

			sort.Strings(expected)
			sort.Strings(existed)

			toolkit.Assert(t, toolkit.Got(nil, existed), toolkit.Want(expected, nil))
		})
	}
}

func TestUnitServerServeStart(t *testing.T) {
	t.Parallel()

	settings := config.MustNew(toolkit.EnvFile())
	settings.Server.Address = "invalid"

	srv := server.New(settings, nil)

	srv.Serve(context.TODO())
}

func TestUnitServerServeShutdown(t *testing.T) {
	t.Parallel()

	settings := config.MustNew(toolkit.EnvFile())
	settings.Server.ShutdownTimeout = 0

	srv := server.New(settings, nil)
	srv.UnsafeRouter().GET("/", echoHandler)

	// simulate unexpected behavior if context will be canceled in the middle of operation
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Millisecond)
	defer cancel()

	go srv.Serve(ctx)
	go func() {
		url := fmt.Sprintf("http://%s/", settings.Server.Address)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

		for range 5 {
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				time.Sleep(time.Millisecond)

				continue
			}

			_ = resp.Body.Close()

			break
		}
	}()

	<-ctx.Done()
}
