package main

import (
	"log"
	"path"

	"github.com/kxnes/go-interviews/apicache/internal/apicache"
	"github.com/kxnes/go-interviews/apicache/internal/drivers/memcache"
	"github.com/kxnes/go-interviews/apicache/internal/drivers/redis"
	"github.com/kxnes/go-interviews/apicache/internal/fs"
	"github.com/kxnes/go-interviews/apicache/internal/options"
)

func main() {
	opts := options.Load(path.Join("configs", "dev.json"))

	var driver fs.Driver

	switch d := opts.Driver; d.Name {
	case "redis":
		driver = redis.New(d.Addr)
	case "memcache":
		driver = memcache.New(d.Addr)
	default:
		log.Fatalln("driver not set")
	}

	srv := apicache.NewServer(
		&apicache.Dependencies{
			Driver: fs.New(driver, opts.FileSystem),
		},
		opts.APICache,
	)

	srv.Listen()
}
