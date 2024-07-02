package options

import (
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/kxnes/go-interviews/apicache/internal/apicache"
	"github.com/kxnes/go-interviews/apicache/internal/fs"
)

// Optional contains optional parameters, like key-value storage
// name from `internal/drivers` package.
// The list of drivers can be expanded with putting `fs.Driver` directly.
type Optional struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
}

// Options contains `Driver` must have parameters.
type Options struct {
	APICache   *apicache.Options `json:"apicache"`
	FileSystem *fs.Options       `json:"filesystem"`
	Driver     *Optional         `json:"driver"`
}

// Load loads config from `p` and parse it to `Options`.
// Panics if something went wrong.
func Load(p string) *Options {
	wd, err := os.Getwd()
	if err != nil {
		log.Panicf("working directory error (%v)", err)
	}

	cfg := path.Join(wd, p)

	f, err := os.Open(cfg)
	if err != nil {
		log.Panicf("open config error (%v)", err)
	}

	opts := &Options{}

	err = json.NewDecoder(f).Decode(opts)
	if err != nil {
		log.Panicf("decode config error (%v)", err)
	}

	return opts
}
