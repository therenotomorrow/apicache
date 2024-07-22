package server

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"

	apiv1delete "github.com/kxnes/go-interviews/apicache/internal/api/v1/delete"
	apiv1get "github.com/kxnes/go-interviews/apicache/internal/api/v1/get"
	apiv1post "github.com/kxnes/go-interviews/apicache/internal/api/v1/post"
	"github.com/kxnes/go-interviews/apicache/internal/config"
	"github.com/kxnes/go-interviews/apicache/pkg/cache"
	"github.com/kxnes/go-interviews/apicache/tools/swagger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Server struct {
	router   *echo.Echo
	settings *config.Settings
}

func New(settings *config.Settings, cache *cache.Cache) *Server {
	router := echo.New()

	router.Debug = settings.Debug
	router.HideBanner = true
	router.HidePort = true

	router.Logger.SetLevel(log.INFO)

	router.Use(middleware.Logger())
	router.Use(middleware.Recover())

	router.GET("/api/v1/:key/", apiv1get.Get(cache))
	router.POST("/api/v1/:key/", apiv1post.Post(cache))
	router.DELETE("/api/v1/:key/", apiv1delete.Delete(cache))

	swagger.Connect(router)

	return &Server{router: router, settings: settings}
}

func (s *Server) UnsafeRouter() *echo.Echo {
	return s.router
}

func (s *Server) Settings() config.Settings {
	return *s.settings
}

func (s *Server) Serve(ctx context.Context) {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := s.router.Start(s.settings.Server.Address); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.router.Logger.Error(err)
			stop()
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(ctx, s.settings.Server.ShutdownTimeout)
	defer cancel()

	if err := s.router.Shutdown(ctx); err != nil {
		s.router.Logger.Error(err)

		_ = s.router.Close()
	}
}
