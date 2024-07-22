//go:build !test

package main

import (
	"context"

	"github.com/kxnes/go-interviews/apicache/internal/config"
	"github.com/kxnes/go-interviews/apicache/internal/server"
	"github.com/kxnes/go-interviews/apicache/pkg/cache"
	"github.com/kxnes/go-interviews/apicache/pkg/drivers/machine"
	"github.com/kxnes/go-interviews/apicache/pkg/drivers/memcached"
	"github.com/kxnes/go-interviews/apicache/pkg/drivers/redis"
)

// @Title            apicache
// @Version          0.0.1
// @Contact.name     Mute Team
// @Contact.url      https://github.com/therenotomorrow/apicache
// @Contact.email    kkxnes@gmail.com
// @Tag.name         cache
// @License.name     MIT
// @License.url      https://github.com/therenotomorrow/apicache/blob/master/LICENSE
func main() {
	var (
		settings = config.MustNew()
		cacher   = cache.MustNew(
			cache.Config{
				MaxConn:     settings.Driver.MaxConn,
				ConnTimeout: settings.Driver.ConnTimeout,
			},
			driver(settings),
		)
	)

	defer func() { _ = cacher.Close() }()

	app := server.New(settings, cacher)

	app.Serve(context.Background())
}

func driver(settings *config.Settings) cache.Driver { //nolint:ireturn
	var driver cache.Driver

	switch drive := settings.Driver; drive.Name {
	case config.DriverMachine:
		driver = machine.New()
	case config.DriverMemcached:
		driver = memcached.NewWithConfig(memcached.Config{Addr: drive.Address})
	case config.DriverRedis:
		driver = redis.NewWithConfig(redis.Config{Addr: drive.Address})
	}

	return driver
}
