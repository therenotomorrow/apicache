package main

import (
	"context"

	"github.com/therenotomorrow/apicache/internal/config"
	"github.com/therenotomorrow/apicache/internal/server"
	"github.com/therenotomorrow/apicache/internal/services/cache"
	"github.com/therenotomorrow/apicache/pkg/drivers/machine"
	"github.com/therenotomorrow/apicache/pkg/drivers/memcached"
	"github.com/therenotomorrow/apicache/pkg/drivers/redis"
)

// @Title            apicache
// @Version          0.0.2
// @Contact.name     Mute Team
// @Contact.url      https://github.com/therenotomorrow/apicache
// @Contact.email    kkxnes@gmail.com
// @Tag.name         cache
// @License.name     MIT
// @License.url      https://github.com/therenotomorrow/apicache/blob/master/LICENSE
func main() {
	var (
		settings *config.Settings
		driver   cache.Driver
		service  *cache.Cache
	)

	settings = config.MustNew()

	switch drive := settings.Driver; drive.Name {
	case config.DriverMachine:
		driver = machine.New()
	case config.DriverMemcached:
		driver = memcached.NewWithConfig(memcached.Config{Addr: drive.Address})
	case config.DriverRedis:
		driver = redis.NewWithConfig(redis.Config{Addr: drive.Address})
	}

	service = cache.MustNew(cache.Config{
		MaxConn:     settings.Driver.MaxConn,
		ConnTimeout: settings.Driver.ConnTimeout,
	}, driver)

	defer func() { _ = service.Close() }()

	app := server.New(settings, service)

	app.Serve(context.Background())
}
