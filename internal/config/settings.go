package config

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type Driver string

const (
	DriverMachine   Driver = "machine"
	DriverMemcached Driver = "memcached"
	DriverRedis     Driver = "redis"
)

var ErrInvalidDriver = errors.New("invalid driver")

type Settings struct {
	Debug  bool `env:"APICACHE_DEBUG,required" json:"debug"`
	Server struct {
		Address         string        `json:"address"`
		ShutdownTimeout time.Duration `json:"shutdownTimeout"`
	} `json:"server"`
	Driver struct {
		Name        Driver        `env:"APICACHE_DRIVER_NAME,required"         json:"name"`
		Address     string        `env:"APICACHE_DRIVER_ADDRESS,required"      json:"address"`
		MaxConn     int           `env:"APICACHE_DRIVER_MAX_CONN,required"     json:"maxConn"`
		ConnTimeout time.Duration `env:"APICACHE_DRIVER_CONN_TIMEOUT,required" json:"connTimeout"`
	} `json:"driver"`
}

func New(filenames ...string) (*Settings, error) {
	settings := new(Settings)

	_ = godotenv.Load(filenames...)

	err := envconfig.Process(context.Background(), settings)
	if err != nil {
		return settings, fmt.Errorf("parse env error: %w", err)
	}

	settings.Server.Address = "0.0.0.0:8080"
	settings.Server.ShutdownTimeout = time.Second

	switch settings.Driver.Name {
	case DriverMachine, DriverMemcached, DriverRedis:
	default:
		return nil, ErrInvalidDriver
	}

	return settings, nil
}

func MustNew(filenames ...string) *Settings {
	settings, err := New(filenames...)
	if err != nil {
		panic(err)
	}

	return settings
}
